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

	// AuthType controls how the credential is sent: bearer | api_key | none.
	AuthType   string `json:"authType,omitempty"`
	Credential string `json:"credential,omitempty"`
	// AuthHeader overrides the default header (Authorization / X-API-Key).
	AuthHeader string `json:"authHeader,omitempty"`
	// AuthHeaderPrefix overrides the default value prefix (bearer's "Bearer ").
	AuthHeaderPrefix string `json:"authHeaderPrefix,omitempty"`
	// Headers are extra non-secret request headers.
	Headers map[string]string `json:"headers,omitempty"`
}

type McpVerifyState struct {
	Status     string `json:"status,omitempty"`
	VerifiedAt string `json:"verifiedAt,omitempty"`
	Error      string `json:"error,omitempty"`
	ToolCount  int    `json:"toolCount"`
	Truncated  bool   `json:"truncated,omitempty"`
}

type McpOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Updated   string `json:"updated,omitempty"`

	Type           string         `json:"type"`
	AuthType       string         `json:"authType"`
	EndpointURL    string         `json:"endpointUrl"`
	Slug           string         `json:"slug"`
	CatalogID      string         `json:"catalogId,omitempty"`
	Status         string         `json:"status,omitempty"` // internal only
	Verify         McpVerifyState `json:"verify"`
	AttachedAgents []string       `json:"attachedAgents,omitempty"`
}

type DescribeMcpResponse struct {
	McpOutput

	Transport        string            `json:"transport"`
	AuthHeader       string            `json:"authHeader,omitempty"`
	AuthHeaderPrefix string            `json:"authHeaderPrefix,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	HasCredential    bool              `json:"hasCredential"`
	SecretRefs       []SecretRef       `json:"secretRefs,omitempty"`
	// Tool count is in Verify.ToolCount; the full list comes from GetMcpTools.
}

// McpToolsResponse is the mcp's cached tool list plus verify state, from GET
// .../mcps/{name}/tools.
type McpToolsResponse struct {
	Status          string           `json:"status,omitempty"`
	VerifiedAt      string           `json:"verifiedAt,omitempty"`
	Error           string           `json:"error,omitempty"`
	ProtocolVersion string           `json:"protocolVersion,omitempty"`
	Tools           []map[string]any `json:"tools"`
	Truncated       bool             `json:"truncated,omitempty"`
	// Names-level diff vs the previous verify snapshot.
	ToolsAdded          []string `json:"toolsAdded,omitempty"`
	ToolsRemoved        []string `json:"toolsRemoved,omitempty"`
	ChangedFromRevision string   `json:"changedFromRevision,omitempty"`
}

type listMcpsResponse struct {
	Mcps []McpOutput `json:"mcps"`
}

type RunMcpToolResult struct {
	Mcp  string `json:"mcp"`
	Tool string `json:"tool"`
	// Result is kept as raw JSON: a tool may return an object, array, or
	// scalar at the top level, and decoding into a typed map would silently
	// fail on the non-object shapes.
	Result json.RawMessage `json:"result,omitempty"`
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
	respBody, err := c.sendMcpRequest(
		ctx,
		http.MethodPost,
		mcpsPath(orgId, projectId, mcpName),
		body,
	)
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
	respBody, err := c.sendMcpRequest(
		ctx,
		http.MethodPut,
		mcpsPath(orgId, projectId, mcpName),
		body,
	)
	if err != nil {
		return "", err
	}
	var result map[string]any
	if jsonErr := json.Unmarshal(respBody, &result); jsonErr == nil {
		if restarted, ok := result["restarted"].([]any); ok && len(restarted) > 0 {
			names := make([]string, 0, len(restarted))
			for _, r := range restarted {
				if name, ok := r.(string); ok {
					names = append(names, name)
				}
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
func (c *DeploymentClient) DescribeMcp(
	ctx context.Context,
	orgId, projectId, mcpName string,
) (*DescribeMcpResponse, error) {
	path := mcpsPath(orgId, projectId, mcpName)
	reqHTTP, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

// GetMcpTools returns the mcp's cached tool list and verify state.
func (c *DeploymentClient) GetMcpTools(
	ctx context.Context,
	orgId, projectId, mcpName string,
) (*McpToolsResponse, error) {
	respBody, err := c.sendMcpRequest(
		ctx, http.MethodGet, mcpsPath(orgId, projectId, mcpName)+"/tools", nil,
	)
	if err != nil {
		return nil, err
	}
	var result McpToolsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode mcp tools response: %w", err)
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

// ListMcpRevisions returns past revisions of an mcp, newest-first.
func (c *DeploymentClient) ListMcpRevisions(
	ctx context.Context,
	orgId, projectId, mcpName string,
) ([]RevisionMeta, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/mcps/%s/revisions",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(mcpName),
	)
	return c.fetchRevisions(ctx, path, "mcp revisions")
}

// DescribeMcpRevision returns the config snapshot of one mcp revision. Decoded
// as a plain map so PrintRevisionDiff can diff any pair of revisions directly.
func (c *DeploymentClient) DescribeMcpRevision(
	ctx context.Context,
	orgId, projectId, mcpName string,
	revision int,
) (map[string]any, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/mcps/%s/revisions/%d",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(mcpName),
		revision,
	)
	respBody, err := c.sendMcpRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode mcp revision response: %w", err)
	}
	return result, nil
}
