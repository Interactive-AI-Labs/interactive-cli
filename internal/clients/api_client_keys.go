package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/auth"
)

type ProjectAPIKey struct {
	ID               string  `json:"id"`
	CreatedAt        string  `json:"createdAt,omitempty"`
	ExpiresAt        *string `json:"expiresAt,omitempty"`
	LastUsedAt       *string `json:"lastUsedAt,omitempty"`
	Note             *string `json:"note,omitempty"`
	PublicKey        string  `json:"publicKey"`
	DisplaySecretKey string  `json:"displaySecretKey"`
	SecretKey        string  `json:"secretKey,omitempty"`
}

type CreateProjectAPIKeyBody struct {
	Note string `json:"note,omitempty"`
}

type RouterAPIKey struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	KeyPreview     string  `json:"key_preview"`
	ProjectID      string  `json:"project_id"`
	UserID         string  `json:"user_id"`
	Disabled       bool    `json:"disabled"`
	Limit          any     `json:"limit,omitempty"`
	LimitRemaining any     `json:"limit_remaining,omitempty"`
	LimitReset     *string `json:"limit_reset,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      *string `json:"updated_at,omitempty"`
	LastUsedAt     *string `json:"last_used_at,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
}

type RouterAPIKeyListResponse struct {
	Keys   []RouterAPIKey `json:"keys"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type CreateRouterAPIKeyBody struct {
	Name               string  `json:"name"`
	Description        string  `json:"description,omitempty"`
	Limit              float64 `json:"limit,omitempty"`
	LimitReset         string  `json:"limit_reset,omitempty"`
	IncludeBYOKInLimit bool    `json:"include_byok_in_limit"`
	ExpiresAt          string  `json:"expires_at,omitempty"`
}

type CreateRouterAPIKeyResponse struct {
	RouterAPIKey
	Key     string `json:"key"`
	Warning string `json:"warning,omitempty"`
}

type DeleteAPIKeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *APIClient) requireKeyManagementAuth() error {
	if c.apiKey != "" || (c.token == "" && len(c.cookies) == 0) {
		return auth.KeyManagementLoginRequiredError()
	}

	return nil
}

func projectAPIKeysPath(orgID, projectID string) string {
	return fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s/api-keys",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
	)
}

func projectAPIKeyPath(orgID, projectID, keyID string) string {
	return fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s/api-keys/%s",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
		url.PathEscape(keyID),
	)
}

func (c *APIClient) ListProjectAPIKeys(
	ctx context.Context,
	orgID, projectID string,
) ([]ProjectAPIKey, error) {
	if err := c.requireKeyManagementAuth(); err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, projectAPIKeysPath(orgID, projectID))
	if err != nil {
		return nil, err
	}

	body, err := c.doAndRead(req, "list project API keys")
	if err != nil {
		return nil, err
	}
	return decodeSuccess[[]ProjectAPIKey](body, "list project API keys")
}

func (c *APIClient) CreateProjectAPIKey(
	ctx context.Context,
	orgID, projectID string,
	body CreateProjectAPIKeyBody,
) (ProjectAPIKey, error) {
	if err := c.requireKeyManagementAuth(); err != nil {
		return ProjectAPIKey{}, err
	}

	payload := map[string]string{}
	if body.Note != "" {
		payload["note"] = body.Note
	}
	req, err := c.newJSONRequest(
		ctx,
		http.MethodPost,
		projectAPIKeysPath(orgID, projectID),
		payload,
	)
	if err != nil {
		return ProjectAPIKey{}, err
	}

	respBody, err := c.doAndRead(req, "create project API key")
	if err != nil {
		return ProjectAPIKey{}, err
	}
	return decodeSuccess[ProjectAPIKey](respBody, "create project API key")
}

func (c *APIClient) DeleteProjectAPIKey(
	ctx context.Context,
	orgID, projectID, keyID string,
) (DeleteAPIKeyResponse, error) {
	if err := c.requireKeyManagementAuth(); err != nil {
		return DeleteAPIKeyResponse{}, err
	}

	req, err := c.newRequest(ctx, http.MethodDelete, projectAPIKeyPath(orgID, projectID, keyID))
	if err != nil {
		return DeleteAPIKeyResponse{}, err
	}

	body, err := c.doAndRead(req, "delete project API key")
	if err != nil {
		return DeleteAPIKeyResponse{}, err
	}
	return decodeSuccess[DeleteAPIKeyResponse](body, "delete project API key")
}

func (c *APIClient) ListRouterAPIKeys(
	ctx context.Context,
	projectID string,
) (RouterAPIKeyListResponse, error) {
	if err := c.requireKeyManagementAuth(); err != nil {
		return RouterAPIKeyListResponse{}, err
	}

	path := fmt.Sprintf("/api/v1/projects/%s/openrouter-keys", url.PathEscape(projectID))
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return RouterAPIKeyListResponse{}, err
	}
	body, err := c.doAndRead(req, "list router keys")
	if err != nil {
		return RouterAPIKeyListResponse{}, err
	}
	var res RouterAPIKeyListResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return RouterAPIKeyListResponse{}, fmt.Errorf("failed to decode router keys: %w", err)
	}
	return res, nil
}

func (c *APIClient) CreateRouterAPIKey(
	ctx context.Context,
	projectID string,
	body CreateRouterAPIKeyBody,
) (CreateRouterAPIKeyResponse, error) {
	if err := c.requireKeyManagementAuth(); err != nil {
		return CreateRouterAPIKeyResponse{}, err
	}

	path := fmt.Sprintf("/api/v1/projects/%s/openrouter-keys", url.PathEscape(projectID))
	req, err := c.newJSONRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return CreateRouterAPIKeyResponse{}, err
	}
	respBody, err := c.doAndRead(req, "create router key")
	if err != nil {
		return CreateRouterAPIKeyResponse{}, err
	}
	var res CreateRouterAPIKeyResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return CreateRouterAPIKeyResponse{}, fmt.Errorf("failed to decode router key: %w", err)
	}
	return res, nil
}

func (c *APIClient) DeleteRouterAPIKey(
	ctx context.Context,
	projectID, keyID string,
) (DeleteAPIKeyResponse, error) {
	if err := c.requireKeyManagementAuth(); err != nil {
		return DeleteAPIKeyResponse{}, err
	}

	path := fmt.Sprintf(
		"/api/v1/projects/%s/openrouter-keys/%s",
		url.PathEscape(projectID),
		url.PathEscape(keyID),
	)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return DeleteAPIKeyResponse{}, err
	}
	body, err := c.doAndRead(req, "delete router key")
	if err != nil {
		return DeleteAPIKeyResponse{}, err
	}
	var res DeleteAPIKeyResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return DeleteAPIKeyResponse{}, fmt.Errorf("failed to decode delete response: %w", err)
	}
	return res, nil
}
