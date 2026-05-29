package clients

import (
	"context"
	"fmt"
	"net/url"
)

// --- Read shapes (mirror backend api/platform/v1/schemas/mcp_connections.py).
// The credential is never returned by the API, so it has no field here.

type ConnectedAgentRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type McpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
	Enabled     bool           `json:"enabled"`
}

type McpConnection struct {
	ID              string              `json:"id"`
	ProjectID       string              `json:"project_id"`
	CatalogID       string              `json:"catalog_id,omitempty"`
	Type            string              `json:"type"`
	Name            string              `json:"name"`
	Slug            string              `json:"slug"`
	Description     string              `json:"description,omitempty"`
	EndpointURL     string              `json:"endpoint_url"`
	Transport       string              `json:"transport"`
	AuthType        string              `json:"auth_type"`
	HasCredential   bool                `json:"has_credential"`
	CustomHeaders   map[string]string   `json:"custom_headers"`
	Status          string              `json:"status"`
	LastVerifiedAt  string              `json:"last_verified_at,omitempty"`
	LastErrorClass  string              `json:"last_error_class,omitempty"`
	ProtocolVersion string              `json:"protocol_version,omitempty"`
	ToolCount       int                 `json:"tool_count"`
	ConnectedAgents []ConnectedAgentRef `json:"connected_agents"`
	CreatedBy       string              `json:"created_by"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

type McpConnectionDetail struct {
	McpConnection
	Tools     []McpTool `json:"tools"`
	LastError string    `json:"last_error,omitempty"`
}

type McpConnectionListData struct {
	Connections []McpConnection `json:"connections"`
}

type McpConnectionDetailData struct {
	Connection McpConnectionDetail `json:"connection"`
}

type McpVerifyData struct {
	Status          string         `json:"status"`
	ErrorClass      string         `json:"error_class,omitempty"`
	ErrorMessage    string         `json:"error_message,omitempty"`
	ProtocolVersion string         `json:"protocol_version,omitempty"`
	ServerInfo      map[string]any `json:"server_info,omitempty"`
	Tools           []McpTool      `json:"tools"`
}

type McpToolCallData struct {
	Status       string         `json:"status"`
	Result       map[string]any `json:"result,omitempty"`
	ErrorClass   string         `json:"error_class,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
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
}

type McpCatalogListData struct {
	Entries []McpCatalogEntry `json:"entries"`
}

// --- Create request. One body covers both connection types; the two
// create-* commands populate it differently.

type McpConnectionCreateBody struct {
	Type          string            `json:"type"`
	CatalogID     string            `json:"catalog_id,omitempty"`
	Name          string            `json:"name"`
	Slug          string            `json:"slug,omitempty"`
	Description   string            `json:"description,omitempty"`
	EndpointURL   string            `json:"endpoint_url,omitempty"`
	Transport     string            `json:"transport,omitempty"`
	AuthType      string            `json:"auth_type"`
	Credential    string            `json:"credential,omitempty"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// mcpBasePath builds the org+project-scoped platform prefix, mirroring
// evalBasePath in api_client_eval.go.
func mcpBasePath(orgID, projectID string) string {
	return fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
	)
}

// --- Read methods (List, Get, Catalog)

func (c *APIClient) ListMcpConnections(
	ctx context.Context, orgID, projectID string,
) (*McpConnectionListData, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections"
	data, _, err := doGet[McpConnectionListData](c, ctx, path, "list mcp connections")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *APIClient) GetMcpConnection(
	ctx context.Context, orgID, projectID, id string,
) (*McpConnectionDetail, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" + url.PathEscape(id)
	data, _, err := doGet[McpConnectionDetailData](c, ctx, path, "get mcp connection")
	if err != nil {
		return nil, err
	}
	return &data.Connection, nil
}

func (c *APIClient) ListMcpCatalog(
	ctx context.Context, orgID, projectID string,
) (*McpCatalogListData, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-catalog"
	data, _, err := doGet[McpCatalogListData](c, ctx, path, "list mcp catalog")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// --- Write methods (Create, Delete, Verify, RunTool)
// These deliberately do NOT call requireAPIKeyMode() — the backend MCP
// endpoints accept session-cookie and bearer-token auth with RBAC, and the
// copilot authenticates with a bearer JWT. Gating on API-key mode would break
// both iai-login (cookie) users and the copilot.
//
// Create/Verify/RunTool discard the raw json.RawMessage returned by doCreate.
// Integrations commands render typed output only and have no raw/JSON output
// mode, so surfacing the raw body would be an unused return value. If a
// raw/JSON output mode is added in the future, thread the second return value
// through here and expose it in the command layer.

func (c *APIClient) CreateMcpConnection(
	ctx context.Context, orgID, projectID string, body McpConnectionCreateBody,
) (*McpConnectionDetail, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections"
	data, _, err := doCreate[McpConnectionDetailData](c, ctx, path, body, "create mcp connection")
	if err != nil {
		return nil, err
	}
	return &data.Connection, nil
}

func (c *APIClient) DeleteMcpConnection(
	ctx context.Context, orgID, projectID, id string,
) error {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" + url.PathEscape(id)
	_, err := c.doDelete(ctx, path, "delete mcp connection")
	return err
}

func (c *APIClient) VerifyMcpConnection(
	ctx context.Context, orgID, projectID, id string,
) (*McpVerifyData, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" + url.PathEscape(id) + "/verify"
	// POST with no body; doCreate handles a nil body via newJSONRequest.
	data, _, err := doCreate[McpVerifyData](c, ctx, path, nil, "verify mcp connection")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *APIClient) RunMcpTool(
	ctx context.Context, orgID, projectID, id, tool string, arguments map[string]any,
) (*McpToolCallData, error) {
	if arguments == nil {
		arguments = map[string]any{}
	}
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" +
		url.PathEscape(id) + "/tools/" + url.PathEscape(tool) + "/run"
	body := struct {
		Arguments map[string]any `json:"arguments"`
	}{Arguments: arguments}
	data, _, err := doCreate[McpToolCallData](c, ctx, path, body, "run mcp tool")
	if err != nil {
		return nil, err
	}
	return &data, nil
}
