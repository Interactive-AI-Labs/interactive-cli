package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CreateMcpBody mirrors the operator's CreateMcpRequest. type=internal
// deploys an in-cluster MCP server; type=external (or catalogId) dials a
// provider URL directly. Fields are mutually exclusive by type — the server
// validates the combination.
type CreateMcpBody struct {
	Type string `json:"type,omitempty"` // "internal" (default) | "external"

	// internal only
	Port       int         `json:"port,omitempty"`
	Image      ImageSpec   `json:"image,omitempty"`
	Resources  Resources   `json:"resources,omitempty"`
	Env        []EnvVar    `json:"env,omitempty"`
	SecretRefs []SecretRef `json:"secretRefs,omitempty"`

	// external only — endpointUrl and catalogId are mutually exclusive.
	EndpointURL string `json:"endpointUrl,omitempty"`
	CatalogID   string `json:"catalogId,omitempty"`

	Credential string `json:"credential,omitempty"`
}

type McpVerifyState struct {
	Status     string `json:"status,omitempty"`
	VerifiedAt string `json:"verifiedAt,omitempty"`
	Error      string `json:"error,omitempty"`
	Version    string `json:"version,omitempty"`
	ToolCount  int    `json:"toolCount"`
	Truncated  bool   `json:"truncated,omitempty"`
}

type McpOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Updated   string `json:"updated,omitempty"`

	Type           string         `json:"type"`
	EndpointURL    string         `json:"endpointUrl"`
	Slug           string         `json:"slug"`
	CatalogID      string         `json:"catalogId,omitempty"`
	Status         string         `json:"status,omitempty"` // internal only
	Verify         McpVerifyState `json:"verify"`
	AttachedAgents []string       `json:"attachedAgents,omitempty"`
}

type DescribeMcpResponse struct {
	McpOutput

	Transport     string           `json:"transport"`
	HasCredential bool             `json:"hasCredential"`
	SecretRefs    []SecretRef      `json:"secretRefs,omitempty"`
	Tools         []map[string]any `json:"tools"`
	ToolVersions  []string         `json:"toolVersions,omitempty"`
}

type listMcpsResponse struct {
	Mcps []McpOutput `json:"mcps"`
}

type RunMcpToolResult struct {
	Mcp    string         `json:"mcp"`
	Tool   string         `json:"tool"`
	Result map[string]any `json:"result"`
}

type verifyMcpResponse struct {
	Message         string           `json:"message"`
	ToolCount       int              `json:"toolCount"`
	Tools           []map[string]any `json:"tools"`
	ProtocolVersion string           `json:"protocolVersion,omitempty"`
	Truncated       bool             `json:"truncated,omitempty"`
}

func mcpsPath(orgId, projectId, mcpName string) string {
	base := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/mcps",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
	)
	if mcpName == "" {
		return base
	}
	return base + "/" + url.PathEscape(mcpName)
}

func (c *DeploymentClient) sendMcpRequest(
	ctx context.Context,
	method, path string,
	body any,
) ([]byte, error) {
	var reqHTTP *http.Request
	var err error
	if body != nil {
		bodyBytes, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", marshalErr)
		}
		reqHTTP, err = c.newRequest(ctx, method, path)
		if err == nil {
			reqHTTP.Header.Set("Content-Type", "application/json")
			reqHTTP.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		}
	} else {
		reqHTTP, err = c.newRequest(ctx, method, path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("mcp request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("mcp request failed with status %s", resp.Status)
	}
	return respBody, nil
}

// CreateMcp deploys an internal MCP or registers an external one (custom
// endpoint or catalog-backed) in the project namespace.
func (c *DeploymentClient) CreateMcp(
	ctx context.Context,
	orgId, projectId, mcpName string,
	body CreateMcpBody,
) (string, error) {
	respBody, err := c.sendMcpRequest(ctx, http.MethodPost, mcpsPath(orgId, projectId, mcpName), body)
	if err != nil {
		return "", err
	}
	return ExtractServerMessage(respBody), nil
}

// PutMcp fully replaces an MCP's spec — the only update mechanism the
// operator exposes for MCPs (no partial patch). A credential change rotates
// the Secret and restarts the MCP and every attached agent.
func (c *DeploymentClient) PutMcp(
	ctx context.Context,
	orgId, projectId, mcpName string,
	body CreateMcpBody,
) (string, error) {
	respBody, err := c.sendMcpRequest(ctx, http.MethodPut, mcpsPath(orgId, projectId, mcpName), body)
	if err != nil {
		return "", err
	}
	var result map[string]any
	if jsonErr := json.Unmarshal(respBody, &result); jsonErr == nil {
		if restarted, ok := result["restarted"].([]any); ok && len(restarted) > 0 {
			names := make([]string, len(restarted))
			for i, r := range restarted {
				names[i], _ = r.(string)
			}
			msg, _ := result["message"].(string)
			return fmt.Sprintf("%s (restarted: %s)", msg, strings.Join(names, ", ")), nil
		}
	}
	return decodeMcpMessage(respBody), nil
}

// decodeMcpMessage extracts the {message, warning} pair the operator's mcp
// endpoints return on success, appending the warning (e.g. dangling agent
// refs) when present.
func decodeMcpMessage(respBody []byte) string {
	var result map[string]any
	if jsonErr := json.Unmarshal(respBody, &result); jsonErr == nil {
		msg, _ := result["message"].(string)
		if warning, ok := result["warning"].(string); ok && warning != "" {
			return fmt.Sprintf("%s\nWarning: %s", msg, warning)
		}
		return msg
	}
	return ExtractServerMessage(respBody)
}

// DeleteMcp uninstalls the mcp's release. If agents are still attached, the
// operator rejects the delete with a 409 unless force is set, in which case
// it proceeds and returns a warning naming the now-dangling agents.
func (c *DeploymentClient) DeleteMcp(
	ctx context.Context,
	orgId, projectId, mcpName string,
	force bool,
) (string, error) {
	path := mcpsPath(orgId, projectId, mcpName)
	if force {
		path += "?force=true"
	}
	respBody, err := c.sendMcpRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return "", err
	}
	return decodeMcpMessage(respBody), nil
}

func (c *DeploymentClient) ListMcps(
	ctx context.Context,
	orgId, projectId string,
) ([]McpOutput, error) {
	respBody, err := c.sendMcpRequest(ctx, http.MethodGet, mcpsPath(orgId, projectId, ""), nil)
	if err != nil {
		return nil, err
	}
	var result listMcpsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode mcps response: %w", err)
	}
	return result.Mcps, nil
}

// DescribeMcp returns the mcp's record, verify state, and cached tools.
// version, when non-empty, reads that image-tag's tools snapshot instead of
// the latest verify (internal mcps only — see `iai mcps get --version`).
func (c *DeploymentClient) DescribeMcp(
	ctx context.Context,
	orgId, projectId, mcpName, version string,
) (*DescribeMcpResponse, error) {
	path := mcpsPath(orgId, projectId, mcpName)
	reqHTTP, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if version != "" {
		q := reqHTTP.URL.Query()
		q.Set("version", version)
		reqHTTP.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("mcp describe request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("mcp describe failed with status %s", resp.Status)
	}

	var result DescribeMcpResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode mcp response: %w", err)
	}
	return &result, nil
}

// VerifyMcp re-dials the MCP and refreshes its cached tools.
func (c *DeploymentClient) VerifyMcp(
	ctx context.Context,
	orgId, projectId, mcpName string,
) (*verifyMcpResponse, error) {
	respBody, err := c.sendMcpRequest(
		ctx, http.MethodPost, mcpsPath(orgId, projectId, mcpName)+"/verify", nil,
	)
	if err != nil {
		return nil, err
	}
	var result verifyMcpResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode verify response: %w", err)
	}
	return &result, nil
}

// RunMcpTool executes one tool via the operator's MCP client. Works for
// external MCPs from anywhere; internal MCPs need the in-cluster operator.
func (c *DeploymentClient) RunMcpTool(
	ctx context.Context,
	orgId, projectId, mcpName, tool string,
	args map[string]any,
) (*RunMcpToolResult, error) {
	body := map[string]any{"tool": tool, "args": args}
	respBody, err := c.sendMcpRequest(
		ctx, http.MethodPost, mcpsPath(orgId, projectId, mcpName)+"/run-tool", body,
	)
	if err != nil {
		return nil, err
	}
	var result RunMcpToolResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode run-tool response: %w", err)
	}
	return &result, nil
}
