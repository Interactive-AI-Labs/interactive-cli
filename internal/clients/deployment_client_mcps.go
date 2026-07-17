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

// CreateMcpBody mirrors the operator's CreateMcpRequest; fields are mutually exclusive by type.
type CreateMcpBody struct {
	Type string `json:"type,omitempty"` // "internal" (default) | "external"

	// internal only
	Port       int         `json:"port,omitempty"`
	Path       string      `json:"path,omitempty"` // default "/mcp"
	Image      ImageSpec   `json:"image,omitempty"`
	Resources  Resources   `json:"resources,omitempty"`
	Env        []EnvVar    `json:"env,omitempty"`
	SecretRefs []SecretRef `json:"secretRefs,omitempty"`

	// external only — endpointUrl and catalogId are mutually exclusive.
	EndpointURL string `json:"endpointUrl,omitempty"`
	CatalogID   string `json:"catalogId,omitempty"`

	// Auth is how the credential is sent.
	Auth McpAuthBody `json:"auth"`
	// Headers are extra non-secret request headers.
	Headers map[string]string `json:"headers,omitempty"`
}

// McpAuthBody is the auth block of a create/update request.
type McpAuthBody struct {
	Type string `json:"type,omitempty"` // bearer | api_key | none
	// Credential is required for bearer and api_key, forbidden for none.
	Credential string `json:"credential,omitempty"`
	// Header overrides the default header (Authorization / X-API-Key).
	Header string `json:"header,omitempty"`
	// HeaderPrefix overrides the default value prefix (bearer's "Bearer ").
	HeaderPrefix string `json:"headerPrefix,omitempty"`
}

// McpAuthInfo describes how an mcp's credential is sent — never the credential.
type McpAuthInfo struct {
	Type         string `json:"type"`
	Header       string `json:"header,omitempty"`
	HeaderPrefix string `json:"headerPrefix,omitempty"`
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
	Auth           McpAuthInfo    `json:"auth"`
	EndpointURL    string         `json:"endpointUrl"`
	Slug           string         `json:"slug"`
	CatalogID      string         `json:"catalogId,omitempty"`
	Status         string         `json:"status,omitempty"` // internal only
	Verify         McpVerifyState `json:"verify"`
	AttachedAgents []string       `json:"attachedAgents,omitempty"`
}

type DescribeMcpResponse struct {
	McpOutput

	Transport     string            `json:"transport"`
	Path          string            `json:"path"`
	Headers       map[string]string `json:"headers,omitempty"`
	HasCredential bool              `json:"hasCredential"`
	SecretRefs    []SecretRef       `json:"secretRefs,omitempty"`
	// Tool count is in Verify.ToolCount; the full list comes from GetMcpTools.
}

// McpToolsResponse is the mcp's cached tool list plus verify state, from GET .../mcps/{name}/tools.
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
	// kept raw: a tool may return an object, array, or scalar at the top level
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

// CreateMcp deploys an internal MCP or registers an external one (custom endpoint or catalog-backed).
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

// PutMcp fully replaces an MCP's spec; a credential change rotates the Secret and restarts the MCP and every attached agent.
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
	return decodeMcpUpdateMessage(respBody), nil
}

// PatchMcp applies a partial update; a credential change rotates the Secret and restarts the MCP and every attached agent.
func (c *DeploymentClient) PatchMcp(
	ctx context.Context,
	orgId, projectId, mcpName string,
	patch UpdatePatch,
) (string, error) {
	respBody, err := c.sendMcpRequest(
		ctx,
		http.MethodPatch,
		mcpsPath(orgId, projectId, mcpName),
		patch,
	)
	if err != nil {
		return "", err
	}
	return decodeMcpUpdateMessage(respBody), nil
}

// decodeMcpUpdateMessage handles PutMcp/PatchMcp's optional {restarted: [...]} on top of the usual {message, warning}.
func decodeMcpUpdateMessage(respBody []byte) string {
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
			return fmt.Sprintf("%s (restarted: %s)", msg, strings.Join(names, ", "))
		}
	}
	return decodeMcpMessage(respBody)
}

// decodeMcpMessage extracts the {message, warning} pair, appending the warning (e.g. dangling agent refs) when present.
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

// DeleteMcp uninstalls the mcp's release; force bypasses the 409 for still-attached agents, returning a dangling-agents warning.
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

// RunMcpTool executes one tool; external MCPs work from anywhere, internal MCPs need the in-cluster operator.
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

// DescribeMcpRevision returns one revision's config snapshot as a plain map, so PrintRevisionDiff can diff any pair directly.
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
