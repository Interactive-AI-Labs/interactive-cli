package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Mirrors the backend's unified MCP API (api/platform/v1/schemas/mcps.py).

type McpBackend string

const (
	McpBackendInternal McpBackend = "internal"
	McpBackendExternal McpBackend = "external"
)

type McpSchema struct {
	Name           string     `json:"name"`
	Backend        McpBackend `json:"backend"`
	Description    *string    `json:"description,omitempty"`
	EndpointURL    *string    `json:"endpoint_url,omitempty"`
	Transport      *string    `json:"transport,omitempty"`
	AuthType       *string    `json:"auth_type,omitempty"`
	HasCredential  bool       `json:"has_credential"`
	CatalogID      *string    `json:"catalog_id,omitempty"`
	Status         *string    `json:"status,omitempty"`
	VerifyStatus   *string    `json:"verify_status,omitempty"`
	ToolCount      int        `json:"tool_count"`
	AttachedAgents []string   `json:"attached_agents"`
	StackId        *string    `json:"stack_id,omitempty"`
}

type McpToolSchema struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

type McpListData struct {
	Mcps              []McpSchema `json:"mcps"`
	InternalAvailable bool        `json:"internal_available"`
}

type McpListResponse struct {
	Data McpListData `json:"data"`
}

type McpDetailResponse struct {
	Data struct {
		Mcp McpSchema `json:"mcp"`
	} `json:"data"`
}

type McpToolsResult struct {
	Data struct {
		Name    string          `json:"name"`
		Backend McpBackend      `json:"backend"`
		Tools   []McpToolSchema `json:"tools"`
	} `json:"data"`
}

type McpVerifyResponse struct {
	Data struct {
		Name       string  `json:"name"`
		Backend    string  `json:"backend"`
		Status     string  `json:"status"`
		ToolCount  *int    `json:"tool_count"`
		ErrorClass *string `json:"error_class,omitempty"`
	} `json:"data"`
}

type McpToolCallResult struct {
	Data struct {
		Name       string          `json:"name"`
		Backend    McpBackend      `json:"backend"`
		Tool       string          `json:"tool"`
		Status     string          `json:"status"`
		Result     json.RawMessage `json:"result"`
		ErrorClass *string         `json:"error_class,omitempty"`
	} `json:"data"`
}

type McpDeleteResponse struct {
	Data struct {
		Name    string `json:"name"`
		Deleted bool   `json:"deleted"`
	} `json:"data"`
}

type McpAuth struct {
	Type         string  `json:"type"`
	Credential   *string `json:"credential,omitempty"`
	HeaderName   *string `json:"header_name,omitempty"`
	HeaderPrefix *string `json:"header_prefix,omitempty"`
}

type McpWorkload struct {
	Image   string `json:"image"`
	Port    int    `json:"port"`
	Path    string `json:"path"`
	Memory  string `json:"memory"`
	CPU     string `json:"cpu"`
	StackId string `json:"stack_id,omitempty"`
}

type McpCreateRequest struct {
	Name        string       `json:"name"`
	Backend     McpBackend   `json:"backend"`
	Description *string      `json:"description,omitempty"`
	CatalogID   *string      `json:"catalog_id,omitempty"`
	EndpointURL *string      `json:"endpoint_url,omitempty"`
	Transport   string       `json:"transport"`
	Auth        McpAuth      `json:"auth"`
	Workload    *McpWorkload `json:"workload,omitempty"`
}

type McpUpdateRequest = map[string]any

// readBody keeps the status code in the error: a 2xx here means the call
// succeeded and only the read dropped, which a bare EOF hides.
func readBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response (HTTP %d): %w", resp.StatusCode, err)
	}
	return body, nil
}

func (c *APIClient) decodeError(body []byte) error {
	var envelope struct {
		Detail struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		} `json:"detail"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Detail.Error.Message != "" {
		return fmt.Errorf("%s: %s", envelope.Detail.Error.Code, envelope.Detail.Error.Message)
	}
	return fmt.Errorf("platform API error: %s", string(body))
}

func (c *APIClient) mcpPath(orgID, projectID string, name string) string {
	path := "/api/platform/v1/organizations/" + orgID + "/projects/" + projectID + "/mcps"
	if name != "" {
		path += "/" + name
	}
	return path
}

type McpCatalogEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	IconKey     string   `json:"icon_key,omitempty"`
	EndpointURL string   `json:"endpoint_url,omitempty"`
	DocsURL     string   `json:"docs_url,omitempty"`
	AuthMethods []string `json:"auth_methods"`
	// Account sign-in details. Absent on static-key entries.
	GrantsAllowed   []string `json:"grants_allowed,omitempty"`
	ScopesSupported []string `json:"scopes_supported,omitempty"`
	// Nil when unknown, so "not established" stays distinct from "false".
	PermissionsAreAMenu *bool `json:"permissions_are_a_menu,omitempty"`
	ScopesSelectable    *bool `json:"scopes_selectable,omitempty"`
	// False means an operator registers an app before anyone can connect.
	// Keyed by grant; absent when unknown, so "not established" stays distinct from "false".
	SelfRegisters map[string]bool `json:"self_registers,omitempty"`
	TokenRenewal  string          `json:"token_renewal,omitempty"`
}

type McpCatalogListData struct {
	Entries []McpCatalogEntry `json:"entries"`
}

func (c *APIClient) ListMcpCatalog(
	ctx context.Context, orgID, projectID string,
) (*McpCatalogListData, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/mcp-catalog"
	data, raw, err := doGet[McpCatalogListData](c, ctx, path, "list mcp catalog")
	if err != nil {
		return nil, nil, err
	}
	return &data, raw, nil
}

func (c *APIClient) ListMcps(
	ctx context.Context,
	orgID, projectID string,
) (*McpListResponse, []byte, error) {
	req, err := c.newRequest(ctx, "GET", c.mcpPath(orgID, projectID, ""))
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpListResponse{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP list: %w", err)
	}
	return out, body, nil
}

func (c *APIClient) DescribeMcp(
	ctx context.Context,
	orgID, projectID, name string,
) (*McpDetailResponse, []byte, error) {
	req, err := c.newRequest(ctx, "GET", c.mcpPath(orgID, projectID, name))
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpDetailResponse{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP detail: %w", err)
	}
	return out, body, nil
}

func (c *APIClient) CreateMcp(
	ctx context.Context,
	orgID, projectID string,
	create McpCreateRequest,
) (*McpDetailResponse, []byte, error) {
	req, err := c.newJSONRequest(ctx, "POST", c.mcpPath(orgID, projectID, ""), create)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 201 {
		return nil, body, c.decodeError(body)
	}
	out := &McpDetailResponse{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP create: %w", err)
	}
	return out, body, nil
}

func (c *APIClient) ListMcpTools(
	ctx context.Context,
	orgID, projectID, name string,
) (*McpToolsResult, []byte, error) {
	req, err := c.newRequest(ctx, "GET", c.mcpPath(orgID, projectID, name)+"/tools")
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpToolsResult{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP tools: %w", err)
	}
	return out, body, nil
}

func (c *APIClient) VerifyMcp(
	ctx context.Context,
	orgID, projectID, name string,
) (*McpVerifyResponse, []byte, error) {
	req, err := c.newRequest(ctx, "POST", c.mcpPath(orgID, projectID, name)+"/verify")
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpVerifyResponse{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP verify: %w", err)
	}
	return out, body, nil
}

func (c *APIClient) RunMcpTool(
	ctx context.Context,
	orgID, projectID, name, tool string,
	arguments map[string]any,
) (*McpToolCallResult, []byte, error) {
	payload := map[string]any{"arguments": arguments}
	req, err := c.newJSONRequest(
		ctx,
		"POST",
		c.mcpPath(orgID, projectID, name)+"/tools/"+tool+"/run",
		payload,
	)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpToolCallResult{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP tool call: %w", err)
	}
	return out, body, nil
}

func (c *APIClient) DeleteMcp(
	ctx context.Context,
	orgID, projectID, name string,
) (*McpDeleteResponse, []byte, error) {
	req, err := c.newRequest(ctx, "DELETE", c.mcpPath(orgID, projectID, name))
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpDeleteResponse{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP delete: %w", err)
	}
	return out, body, nil
}

type McpSignInResult struct {
	// Wrapped, like every other platform response: the URL lives at
	// data.authorization_url, and decoding the envelope flat silently yields
	// an empty string rather than an error.
	Data struct {
		Name         string `json:"name"`
		AuthorizeURL string `json:"authorization_url"`
	} `json:"data"`
}

func (c *APIClient) BeginMcpSignIn(
	ctx context.Context,
	orgID, projectID, name string,
) (*McpSignInResult, []byte, error) {
	req, err := c.newRequest(ctx, "POST", c.mcpPath(orgID, projectID, name)+"/connect")
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpSignInResult{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP sign-in: %w", err)
	}
	return out, body, nil
}

// ConnectionStatus reports whether the project's connection for this MCP holds
// a live provider credential (connected) or not.
func (c *APIClient) ConnectionStatus(
	ctx context.Context,
	orgID, projectID, name string,
) (bool, error) {
	req, err := c.newRequest(ctx, "GET", c.mcpPath(orgID, projectID, name)+"/connection")
	if err != nil {
		return false, err
	}
	resp, err := c.do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, c.decodeError(body)
	}
	var out struct {
		Data struct {
			Connected bool `json:"connected"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return false, fmt.Errorf("failed to decode MCP connection status: %w", err)
	}
	return out.Data.Connected, nil
}

// Disconnect forgets the provider credential stored for this MCP.
func (c *APIClient) Disconnect(ctx context.Context, orgID, projectID, name string) error {
	req, err := c.newRequest(ctx, "DELETE", c.mcpPath(orgID, projectID, name)+"/connection")
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return c.decodeError(body)
	}
	return nil
}

func (c *APIClient) UpdateMcp(
	ctx context.Context,
	orgID, projectID, name string,
	patch map[string]any,
) (*McpDetailResponse, []byte, error) {
	req, err := c.newJSONRequest(ctx, "PATCH", c.mcpPath(orgID, projectID, name), patch)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body, c.decodeError(body)
	}
	out := &McpDetailResponse{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, body, fmt.Errorf("failed to decode MCP update: %w", err)
	}
	return out, body, nil
}
