package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SearchHit is one ranked result.
type SearchHit struct {
	ID       string         `json:"id"`
	Score    float64        `json:"score"`
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata"`
}

// SearchResponse is a single search's ranked results.
type SearchResponse struct {
	Results []SearchHit `json:"results"`
	HasMore bool        `json:"hasMore"`
}

// BatchSearchResponse is one SearchResponse per sub-search.
type BatchSearchResponse struct {
	Responses []SearchResponse `json:"responses"`
}

func searchBase(orgId, projectId, database, collection string) string {
	return CollectionsPath(orgId, projectId, database) + "/" + url.PathEscape(collection)
}

// postSearch POSTs a JSON body and decodes a SearchResponse.
func (c *DeploymentClient) postSearch(
	ctx context.Context,
	path string,
	body []byte,
	action string,
) (*SearchResponse, error) {
	var result SearchResponse
	if err := c.sendJSONInto(ctx, http.MethodPost, path, body, action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Search runs a single-lane search; exact routes to the exhaustive scan.
func (c *DeploymentClient) Search(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
	exact bool,
) (*SearchResponse, error) {
	path := searchBase(orgId, projectId, database, collection) + "/search"
	if exact {
		path += "/exact"
	}
	return c.postSearch(ctx, path, body, "search")
}

// HybridSearch runs a multi-lane (RRF-fused) search. It POSTs to the same
// /search endpoint as Search; the server picks the hybrid path off the
// "mode":"hybrid" discriminator, which this method sets so the caller's file
// never has to include it.
func (c *DeploymentClient) HybridSearch(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
) (*SearchResponse, error) {
	body, err := withHybridMode(body)
	if err != nil {
		return nil, err
	}
	path := searchBase(orgId, projectId, database, collection) + "/search"
	return c.postSearch(ctx, path, body, "hybrid search")
}

// withHybridMode forces "mode":"hybrid" onto a search body. The server keys the
// hybrid path off this discriminator, not the body shape, so a file that omits
// it would otherwise be parsed as a single search and rejected.
func withHybridMode(body []byte) ([]byte, error) {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("hybrid search body must be a JSON object: %w", err)
	}
	m["mode"] = "hybrid"
	return json.Marshal(m)
}

// QueryByID finds neighbors of an existing chunk by its stored vector.
func (c *DeploymentClient) QueryByID(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
) (*SearchResponse, error) {
	path := searchBase(orgId, projectId, database, collection) + "/query-by-id"
	return c.postSearch(ctx, path, body, "query-by-id")
}

// SearchBatch runs several searches in one request.
func (c *DeploymentClient) SearchBatch(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
) (*BatchSearchResponse, error) {
	path := searchBase(orgId, projectId, database, collection) + "/search/batch"
	var result BatchSearchResponse
	if err := c.sendJSONInto(ctx, http.MethodPost, path, body, "batch search", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
