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
	MarketingName string `json:"marketingName"`
	Developer     string `json:"developer"`
	// Backends older than the developer field send the developer only here.
	EndpointProvider string             `json:"endpointProvider"`
	Description      string             `json:"description"`
	ContextLength    *int               `json:"contextLength"`
	Capabilities     []string           `json:"capabilities"`
	Prices           map[string]float64 `json:"prices"`
	TokenizerID      string             `json:"tokenizerId"`
	CreatedAt        string             `json:"createdAt"`
	Region           string             `json:"region"`
	HasOtherRegion   bool               `json:"hasOtherRegion"`
	Recommendations  map[string]int     `json:"recommendations"`
	ProviderNames    []string           `json:"providerNames"`
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
	Models     json.RawMessage `json:"models"`
	TotalCount int             `json:"totalCount"`
}

func (c *APIClient) ListRouterModels(
	ctx context.Context,
	projectID string,
	opts RouterModelListOptions,
) ([]RouterModel, PageMeta, json.RawMessage, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/models/router-models")
	if err != nil {
		return nil, PageMeta{}, nil, fmt.Errorf("failed to create request: %w", err)
	}

	q, err := query.Values(opts)
	if err != nil {
		return nil, PageMeta{}, nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}
	q.Set("projectId", projectID)
	req.URL.RawQuery = q.Encode()

	body, err := c.doAndRead(req, "list router models")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}

	return decodeRouterModelList(body, opts)
}

func decodeRouterModelList(
	body []byte,
	opts RouterModelListOptions,
) ([]RouterModel, PageMeta, json.RawMessage, error) {
	var data routerModelListData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, PageMeta{}, nil, fmt.Errorf(
			"failed to decode list router models response: %w",
			err,
		)
	}
	var models []RouterModel
	if err := json.Unmarshal(data.Models, &models); err != nil {
		return nil, PageMeta{}, nil, fmt.Errorf(
			"failed to decode list router models response: %w",
			err,
		)
	}

	meta := PageMeta{Page: opts.Page + 1, Limit: opts.Limit, TotalItems: data.TotalCount}
	if opts.Limit > 0 {
		meta.TotalPages = (data.TotalCount + opts.Limit - 1) / opts.Limit
	}
	return models, meta, data.Models, nil
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
