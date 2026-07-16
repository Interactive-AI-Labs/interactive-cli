package clients

import (
	"context"
	"encoding/json"
)

// These mirror the backend mcp-catalog schemas.

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
