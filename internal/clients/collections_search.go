package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	return collectionsPath(orgId, projectId, database) + "/" + url.PathEscape(collection)
}

// postSearch POSTs a JSON body and decodes a SearchResponse.
func (c *DeploymentClient) postSearch(
	ctx context.Context,
	path string,
	body []byte,
	action string,
) (*SearchResponse, error) {
	var result SearchResponse
	if err := c.postJSONInto(ctx, path, body, action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// postJSONInto POSTs a JSON body and decodes the response into dst.
func (c *DeploymentClient) postJSONInto(
	ctx context.Context,
	path string,
	body []byte,
	action string,
	dst any,
) error {
	req, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", action, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return collectionErr(resp, action)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("failed to decode %s response: %w", action, err)
	}
	return nil
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

// HybridSearch runs a multi-lane (RRF-fused) search. Note it POSTs to the same
// /search endpoint as Search — the server dispatches to single-lane vs hybrid
// based on the request body (a top-level "queries" array selects hybrid), so the
// `search` and `search hybrid` sub-commands hit the same URL with different bodies.
func (c *DeploymentClient) HybridSearch(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
) (*SearchResponse, error) {
	path := searchBase(orgId, projectId, database, collection) + "/search"
	return c.postSearch(ctx, path, body, "hybrid search")
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
	if err := c.postJSONInto(ctx, path, body, "batch search", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
