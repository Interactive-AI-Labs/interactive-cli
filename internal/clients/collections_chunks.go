package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Chunk is a single stored chunk (a row in a collection).
type Chunk struct {
	ID         string         `json:"id"`
	DocumentID string         `json:"documentId"`
	Text       string         `json:"text"`
	Metadata   map[string]any `json:"metadata"`
	// Vector / Vectors are only populated by Get with includeVector.
	Vector  []float64                  `json:"vector,omitempty"`
	Vectors map[string]json.RawMessage `json:"vectors,omitempty"`
}

// ChunkList is a paginated list of chunks.
type ChunkList struct {
	Chunks     []Chunk `json:"chunks"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

// ChunkResult is one entry in an upsert response.
type ChunkResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ChunkUpsertResult is the response from a batch upsert.
type ChunkUpsertResult struct {
	Results []ChunkResult    `json:"results"`
	Errors  []map[string]any `json:"errors"`
}

// BulkDeleteResult is the response from a bulk delete.
type BulkDeleteResult struct {
	DeletedCount int64    `json:"deletedCount"`
	DeletedIds   []string `json:"deletedIds"`
}

// ListChunksOpts carries the optional list query parameters.
type ListChunksOpts struct {
	Limit  int
	Cursor string
	Prefix string
}

func chunksPath(orgId, projectId, database, collection string) string {
	base := collectionsPath(orgId, projectId, database)
	return base + "/" + url.PathEscape(collection) + "/chunks"
}

// UpsertChunks upserts a batch of chunks from a raw JSON body.
func (c *DeploymentClient) UpsertChunks(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
) (*ChunkUpsertResult, error) {
	path := chunksPath(orgId, projectId, database, collection)
	req, err := c.newRequest(ctx, http.MethodPut, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("upsert chunks request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "upsert chunks")
	}

	var result ChunkUpsertResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode upsert response: %w", err)
	}
	return &result, nil
}

// ListChunks returns a page of chunks.
func (c *DeploymentClient) ListChunks(
	ctx context.Context,
	orgId, projectId, database, collection string,
	opts ListChunksOpts,
) (*ChunkList, error) {
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		chunksPath(orgId, projectId, database, collection),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	if opts.Prefix != "" {
		q.Set("prefix", opts.Prefix)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("list chunks request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "list chunks")
	}

	var result ChunkList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode chunks response: %w", err)
	}
	return &result, nil
}

// GetChunk fetches a single chunk, optionally including its vector(s).
func (c *DeploymentClient) GetChunk(
	ctx context.Context,
	orgId, projectId, database, collection, id string,
	includeVector bool,
) (*Chunk, error) {
	path := chunksPath(orgId, projectId, database, collection) + "/" + url.PathEscape(id)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if includeVector {
		q := req.URL.Query()
		q.Set("include_vector", "true")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("get chunk request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "get chunk")
	}

	var result Chunk
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode chunk response: %w", err)
	}
	return &result, nil
}

// CountChunks returns the chunk count, optionally scoped by a filter/prefix body.
func (c *DeploymentClient) CountChunks(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
) (int64, error) {
	path := chunksPath(orgId, projectId, database, collection) + "/count"
	req, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.do(req)
	if err != nil {
		return 0, fmt.Errorf("count chunks request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, collectionErr(resp, "count chunks")
	}

	var result struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}
	return result.Count, nil
}

// PatchChunk updates a chunk's metadata and/or text from a raw JSON body.
func (c *DeploymentClient) PatchChunk(
	ctx context.Context,
	orgId, projectId, database, collection, id string,
	body []byte,
) (*Chunk, error) {
	path := chunksPath(orgId, projectId, database, collection) + "/" + url.PathEscape(id)
	req, err := c.newRequest(ctx, http.MethodPatch, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("patch chunk request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "patch chunk")
	}

	var result Chunk
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode chunk response: %w", err)
	}
	return &result, nil
}

// DeleteChunk deletes a single chunk by id.
func (c *DeploymentClient) DeleteChunk(
	ctx context.Context,
	orgId, projectId, database, collection, id string,
) (string, error) {
	path := chunksPath(orgId, projectId, database, collection) + "/" + url.PathEscape(id)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("delete chunk request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	msg := ExtractServerMessage(respBody)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", fmt.Errorf("failed to delete chunk: server returned %s", resp.Status)
	}
	return msg, nil
}

// BulkDeleteChunks deletes chunks by ids/filter/all. confirmAll sets the
// X-Confirm-Delete header required by the "all" selector.
func (c *DeploymentClient) BulkDeleteChunks(
	ctx context.Context,
	orgId, projectId, database, collection string,
	body []byte,
	confirmAll bool,
) (*BulkDeleteResult, error) {
	path := chunksPath(orgId, projectId, database, collection)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if confirmAll {
		req.Header.Set("X-Confirm-Delete", "yes")
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("bulk delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "bulk delete chunks")
	}

	var result BulkDeleteResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode bulk delete response: %w", err)
	}
	return &result, nil
}
