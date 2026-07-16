package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

// RouterModel is a model exposed by the router-models endpoints. Unlike the
// eval-scoped resources, these endpoints take projectId as a query parameter
// and return the resource directly (no {success, data} envelope).
type RouterModel struct {
	ID              string             `json:"id"`
	ProjectID       string             `json:"projectId,omitempty"`
	ModelName       string             `json:"modelName"`
	MarketingName   string             `json:"marketingName,omitempty"`
	MatchPattern    string             `json:"matchPattern"`
	Description     string             `json:"description,omitempty"`
	ContextLength   *int               `json:"contextLength,omitempty"`
	Capabilities    []string           `json:"capabilities"`
	Prices          map[string]float64 `json:"prices"`
	TokenizerID     string             `json:"tokenizerId,omitempty"`
	TokenizerConfig json.RawMessage    `json:"tokenizerConfig,omitempty"`
	CreatedAt       string             `json:"createdAt,omitempty"`
	LastUsed        string             `json:"lastUsed,omitempty"`
	Region          string             `json:"region"`
	HasOtherRegion  bool               `json:"hasOtherRegion"`
}

type RouterModelListOptions struct {
	// Page is 0-indexed; omitempty drops page=0 from the query, which the API
	// treats as the default first page.
	Page   int    `url:"page,omitempty"`
	Limit  int    `url:"limit,omitempty"`
	Search string `url:"search,omitempty"`
	Region string `url:"region,omitempty"`
}

type routerModelListData struct {
	Models     []RouterModel `json:"models"`
	TotalCount int           `json:"totalCount"`
}

// ListRouterModels lists router models for a project. This endpoint is
// 0-indexed and returns only a totalCount, so we synthesize a PageMeta (with a
// 1-indexed Page) to match the shape of the eval-scoped list endpoints.
func (c *APIClient) ListRouterModels(
	ctx context.Context,
	projectID string,
	opts RouterModelListOptions,
) ([]RouterModel, PageMeta, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/models/router-models")
	if err != nil {
		return nil, PageMeta{}, fmt.Errorf("failed to create request: %w", err)
	}

	q, err := query.Values(opts)
	if err != nil {
		return nil, PageMeta{}, fmt.Errorf("failed to encode query parameters: %w", err)
	}
	q.Set("projectId", projectID)
	req.URL.RawQuery = q.Encode()

	body, err := c.doAndRead(req, "list router models")
	if err != nil {
		return nil, PageMeta{}, err
	}

	var data routerModelListData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, PageMeta{}, fmt.Errorf(
			"failed to decode list router models response: %w",
			err,
		)
	}

	meta := PageMeta{
		Page:       opts.Page + 1,
		Limit:      opts.Limit,
		TotalItems: data.TotalCount,
	}
	if opts.Limit > 0 {
		meta.TotalPages = (data.TotalCount + opts.Limit - 1) / opts.Limit
	}
	return data.Models, meta, nil
}

// GetRouterModelByID retrieves a single router model by its local UUID.
func (c *APIClient) GetRouterModelByID(
	ctx context.Context,
	projectID, modelID string,
) (*RouterModel, error) {
	path := "/api/v1/models/router-models/by-id/" + url.PathEscape(modelID)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Set("projectId", projectID)
	req.URL.RawQuery = q.Encode()

	body, err := c.doAndRead(req, "get router model")
	if err != nil {
		return nil, err
	}

	var model RouterModel
	if err := json.Unmarshal(body, &model); err != nil {
		return nil, fmt.Errorf("failed to decode router model response: %w", err)
	}
	return &model, nil
}
