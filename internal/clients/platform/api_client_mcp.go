package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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

type McpDetailData struct {
	Mcp McpSchema `json:"mcp"`
}

type McpToolsData struct {
	Name    string          `json:"name"`
	Backend McpBackend      `json:"backend"`
	Tools   []McpToolSchema `json:"tools"`
}

type McpVerifyData struct {
	Name       string     `json:"name"`
	Backend    McpBackend `json:"backend"`
	Status     string     `json:"status"`
	ToolCount  int        `json:"tool_count"`
	ErrorClass *string    `json:"error_class,omitempty"`
}

type McpToolCallData struct {
	Name       string          `json:"name"`
	Backend    McpBackend      `json:"backend"`
	Tool       string          `json:"tool"`
	Status     string          `json:"status"`
	Result     json.RawMessage `json:"result"`
	ErrorClass *string         `json:"error_class,omitempty"`
}

type McpDeleteData struct {
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
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

type McpCatalogEntry struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Category            string          `json:"category"`
	Description         string          `json:"description,omitempty"`
	Type                string          `json:"type"`
	IconKey             string          `json:"icon_key,omitempty"`
	EndpointURL         string          `json:"endpoint_url,omitempty"`
	DocsURL             string          `json:"docs_url,omitempty"`
	AuthMethods         []string        `json:"auth_methods"`
	GrantsAllowed       []string        `json:"grants_allowed,omitempty"`
	ScopesSupported     []string        `json:"scopes_supported,omitempty"`
	PermissionsAreAMenu *bool           `json:"permissions_are_a_menu,omitempty"`
	ScopesSelectable    *bool           `json:"scopes_selectable,omitempty"`
	SelfRegisters       map[string]bool `json:"self_registers,omitempty"`
	TokenRenewal        string          `json:"token_renewal,omitempty"`
}

type McpCatalogListData struct {
	Entries []McpCatalogEntry `json:"entries"`
}

type McpSignInData struct {
	Name         string `json:"name"`
	AuthorizeURL string `json:"authorization_url"`
}

type mcpConnectionData struct {
	Connected bool `json:"connected"`
}

func (c *APIClient) mcpPath(orgID, projectID, name string) string {
	path := evalBasePath(orgID, projectID) + "/mcps"
	if name != "" {
		path += "/" + url.PathEscape(name)
	}
	return path
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

func (c *APIClient) McpCatalogEntry(
	ctx context.Context, orgID, projectID, catalogID string,
) (*McpCatalogEntry, error) {
	catalog, _, err := c.ListMcpCatalog(ctx, orgID, projectID)
	if err != nil {
		return nil, err
	}
	return FindMcpCatalogEntry(catalog.Entries, catalogID)
}

// FindMcpCatalogEntry resolves one entry by id, naming the command that lists them.
func FindMcpCatalogEntry(entries []McpCatalogEntry, catalogID string) (*McpCatalogEntry, error) {
	for i := range entries {
		if entries[i].ID == catalogID {
			return &entries[i], nil
		}
	}
	return nil, fmt.Errorf("no catalog entry %q — see 'iai mcps catalog'", catalogID)
}

func (c *APIClient) ListMcps(
	ctx context.Context, orgID, projectID string,
) (*McpListData, json.RawMessage, error) {
	data, raw, err := doGet[McpListData](c, ctx, c.mcpPath(orgID, projectID, ""), "list mcps")
	if err != nil {
		return nil, nil, err
	}
	return &data, raw, nil
}

func (c *APIClient) DescribeMcp(
	ctx context.Context, orgID, projectID, name string,
) (*McpSchema, json.RawMessage, error) {
	data, raw, err := doGet[McpDetailData](
		c,
		ctx,
		c.mcpPath(orgID, projectID, name),
		"describe mcp",
	)
	if err != nil {
		return nil, nil, err
	}
	return &data.Mcp, raw, nil
}

func (c *APIClient) CreateMcp(
	ctx context.Context, orgID, projectID string, create McpCreateRequest,
) (*McpSchema, json.RawMessage, error) {
	data, raw, err := doCreate[McpDetailData](
		c, ctx, c.mcpPath(orgID, projectID, ""), create, "create mcp",
	)
	if err != nil {
		return nil, nil, err
	}
	return &data.Mcp, raw, nil
}

func (c *APIClient) ListMcpTools(
	ctx context.Context, orgID, projectID, name string,
) (*McpToolsData, json.RawMessage, error) {
	path := c.mcpPath(orgID, projectID, name) + "/tools"
	data, raw, err := doGet[McpToolsData](c, ctx, path, "list mcp tools")
	if err != nil {
		return nil, nil, err
	}
	return &data, raw, nil
}

func (c *APIClient) VerifyMcp(
	ctx context.Context, orgID, projectID, name string,
) (*McpVerifyData, json.RawMessage, error) {
	path := c.mcpPath(orgID, projectID, name) + "/verify"
	data, raw, err := doCreate[McpVerifyData](c, ctx, path, nil, "verify mcp")
	if err != nil {
		return nil, nil, err
	}
	return &data, raw, nil
}

func (c *APIClient) RunMcpTool(
	ctx context.Context,
	orgID, projectID, name, tool string,
	arguments map[string]any,
) (*McpToolCallData, json.RawMessage, error) {
	path := c.mcpPath(orgID, projectID, name) + "/tools/" + url.PathEscape(tool) + "/run"
	data, raw, err := doCreate[McpToolCallData](
		c, ctx, path, map[string]any{"arguments": arguments}, "run mcp tool",
	)
	if err != nil {
		return nil, nil, err
	}
	return &data, raw, nil
}

func (c *APIClient) DeleteMcp(
	ctx context.Context, orgID, projectID, name string,
) (*McpDeleteData, json.RawMessage, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, c.mcpPath(orgID, projectID, name))
	if err != nil {
		return nil, nil, err
	}
	body, err := c.doAndRead(req, "delete mcp")
	if err != nil {
		return nil, nil, err
	}
	data, err := decodeSuccess[McpDeleteData](body, "delete mcp")
	if err != nil {
		return nil, json.RawMessage(body), err
	}
	return &data, json.RawMessage(body), nil
}

func (c *APIClient) BeginMcpSignIn(
	ctx context.Context, orgID, projectID, name string,
) (*McpSignInData, json.RawMessage, error) {
	path := c.mcpPath(orgID, projectID, name) + "/connect"
	data, raw, err := doCreate[McpSignInData](c, ctx, path, nil, "connect mcp")
	if err != nil {
		return nil, nil, err
	}
	return &data, raw, nil
}

func (c *APIClient) ConnectionStatus(
	ctx context.Context, orgID, projectID, name string,
) (bool, error) {
	path := c.mcpPath(orgID, projectID, name) + "/connection"
	data, _, err := doGet[mcpConnectionData](c, ctx, path, "get mcp connection")
	return data.Connected, err
}

func (c *APIClient) Disconnect(ctx context.Context, orgID, projectID, name string) error {
	_, err := c.doDelete(ctx, c.mcpPath(orgID, projectID, name)+"/connection", "disconnect mcp")
	return err
}

func (c *APIClient) UpdateMcp(
	ctx context.Context,
	orgID, projectID, name string,
	patch McpUpdateRequest,
) (*McpSchema, json.RawMessage, error) {
	data, raw, err := doUpdate[McpDetailData](
		c, ctx, c.mcpPath(orgID, projectID, name), patch, "update mcp",
	)
	if err != nil {
		return nil, nil, err
	}
	return &data.Mcp, raw, nil
}
