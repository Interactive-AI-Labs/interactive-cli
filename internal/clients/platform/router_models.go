package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

const (
	RouterModelRegionUS = "us"
	RouterModelRegionEU = "eu"
)

// RouterModel is a model exposed by the router-models endpoints. Unlike the
// eval-scoped resources, these endpoints take projectId as a query parameter
// and return the resource directly (no {success, data} envelope).
type RouterModel struct {
	ID            string `json:"id"`
	ModelName     string `json:"modelName"`
	MarketingName string `json:"marketingName,omitempty"`
	Developer     string `json:"developer,omitempty"`
	// EndpointProvider is the deprecated name of Developer, still the only one older backends send.
	EndpointProvider string             `json:"endpointProvider,omitempty"`
	Description      string             `json:"description,omitempty"`
	ContextLength    *int               `json:"contextLength,omitempty"`
	Capabilities     []string           `json:"capabilities"`
	Prices           map[string]float64 `json:"prices"`
	TokenizerID      string             `json:"tokenizerId,omitempty"`
	TokenizerConfig  json.RawMessage    `json:"tokenizerConfig,omitempty"`
	CreatedAt        string             `json:"createdAt,omitempty"`
	Region           string             `json:"region"`
	HasOtherRegion   bool               `json:"hasOtherRegion"`
	Recommendations  map[string]int     `json:"recommendations,omitempty"`
}

type RouterModelProvider struct {
	Name            string             `json:"name"`
	Tag             *string            `json:"tag"`
	Prices          map[string]float64 `json:"prices"`
	ContextLength   *int               `json:"contextLength"`
	MaxOutputTokens *int               `json:"maxOutputTokens"`
	UptimeLastDay   *float64           `json:"uptimeLastDay"`
}

type RouterModelLatency struct {
	Date                    string   `json:"date"`
	TimeToFirstTokenSeconds *float64 `json:"timeToFirstTokenSeconds"`
	SecondsPerOutputToken   *float64 `json:"secondsPerOutputToken"`
}

type RouterModelDetail struct {
	RouterModel
	Providers []RouterModelProvider `json:"providers"`
	Latency   *RouterModelLatency   `json:"latency"`
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
) (*RouterModelDetail, json.RawMessage, error) {
	path := "/api/v1/models/router-models/by-id/" + url.PathEscape(modelID)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Set("projectId", projectID)
	req.URL.RawQuery = q.Encode()

	body, err := c.doAndRead(req, "get router model")
	if err != nil {
		return nil, nil, err
	}

	var model RouterModelDetail
	if err := json.Unmarshal(body, &model); err != nil {
		return nil, nil, fmt.Errorf("failed to decode router model response: %w", err)
	}
	return &model, json.RawMessage(body), nil
}
