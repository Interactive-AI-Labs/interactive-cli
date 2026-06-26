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
	ProjectID string `json:"projectId"`
	Note      string `json:"note,omitempty"`
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

func (c *APIClient) requireCookieMode() error {
	if c.token != "" || c.apiKey != "" || len(c.cookies) == 0 {
		return auth.KeyManagementLoginRequiredError()
	}

	return nil
}

func (c *APIClient) trpcMutation(ctx context.Context, path string, input any, out any) error {
	if err := c.requireCookieMode(); err != nil {
		return err
	}

	req, err := c.newJSONRequest(
		ctx,
		http.MethodPost,
		"/api/trpc/"+path,
		map[string]any{"json": input},
	)
	if err != nil {
		return err
	}

	body, err := c.doAndRead(req, path)
	if err != nil {
		return err
	}

	return decodeTRPCData(body, out)
}

func (c *APIClient) trpcQuery(ctx context.Context, path string, input any, out any) error {
	if err := c.requireCookieMode(); err != nil {
		return err
	}

	encoded, err := json.Marshal(map[string]any{"json": input})
	if err != nil {
		return fmt.Errorf("failed to encode tRPC input: %w", err)
	}
	u, err := url.Parse(c.hostname)
	if err != nil {
		return fmt.Errorf("failed to parse API hostname: %w", err)
	}
	u.Path = "/api/trpc/" + path
	u.RawQuery = "input=" + url.QueryEscape(string(encoded))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	body, err := c.doAndRead(req, path)
	if err != nil {
		return err
	}

	return decodeTRPCData(body, out)
}

func decodeTRPCData(body []byte, out any) error {
	var envelope struct {
		Result struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("failed to decode tRPC response: %w", err)
	}
	if envelope.Error != nil {
		if msg := ExtractServerMessage(body); msg != "" {
			return fmt.Errorf("tRPC error: %s", msg)
		}
		return fmt.Errorf("tRPC error")
	}
	if len(envelope.Result.Data) == 0 {
		return nil
	}

	var wrapped struct {
		JSON json.RawMessage `json:"json"`
	}
	if err := json.Unmarshal(envelope.Result.Data, &wrapped); err == nil && len(wrapped.JSON) > 0 {
		return json.Unmarshal(wrapped.JSON, out)
	}
	return json.Unmarshal(envelope.Result.Data, out)
}

func (c *APIClient) ListProjectAPIKeys(
	ctx context.Context,
	projectID string,
) ([]ProjectAPIKey, error) {
	var keys []ProjectAPIKey
	err := c.trpcQuery(
		ctx,
		"projectApiKeys.byProjectId",
		map[string]string{"projectId": projectID},
		&keys,
	)
	return keys, err
}

func (c *APIClient) CreateProjectAPIKey(
	ctx context.Context,
	body CreateProjectAPIKeyBody,
) (ProjectAPIKey, error) {
	var key ProjectAPIKey
	err := c.trpcMutation(ctx, "projectApiKeys.create", body, &key)
	return key, err
}

func (c *APIClient) DeleteProjectAPIKey(
	ctx context.Context,
	projectID, keyID string,
) (DeleteAPIKeyResponse, error) {
	var ok bool
	if err := c.trpcMutation(
		ctx,
		"projectApiKeys.delete",
		map[string]string{"projectId": projectID, "id": keyID},
		&ok,
	); err != nil {
		return DeleteAPIKeyResponse{}, err
	}
	return DeleteAPIKeyResponse{Success: ok}, nil
}

func (c *APIClient) ListRouterAPIKeys(
	ctx context.Context,
	projectID string,
) (RouterAPIKeyListResponse, error) {
	if err := c.requireCookieMode(); err != nil {
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
	if err := c.requireCookieMode(); err != nil {
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
	if err := c.requireCookieMode(); err != nil {
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
