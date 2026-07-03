package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// DocumentSummary is one entry in a documents list (chunks grouped by documentId).
type DocumentSummary struct {
	DocumentID string `json:"documentId"`
	ChunkCount int64  `json:"chunkCount"`
}

// DocumentList is a paginated list of documents.
type DocumentList struct {
	Documents  []DocumentSummary `json:"documents"`
	NextCursor *string           `json:"nextCursor"`
	HasMore    bool              `json:"hasMore"`
}

// DocumentChunks is a single document's chunks (paginated).
type DocumentChunks struct {
	DocumentID string  `json:"documentId"`
	Chunks     []Chunk `json:"chunks"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

// DeleteDocumentResult is the response from deleting a document.
type DeleteDocumentResult struct {
	DocumentID   string `json:"documentId"`
	DeletedCount int64  `json:"deletedCount"`
}

func documentsPath(orgId, projectId, database, collection string) string {
	base := CollectionsPath(orgId, projectId, database)
	return base + "/" + url.PathEscape(collection) + "/documents"
}

// ListDocuments returns a page of documents (chunks grouped by documentId).
func (c *DeploymentClient) ListDocuments(
	ctx context.Context,
	orgId, projectId, database, collection string,
	limit int,
	cursor string,
	filter string,
) (*DocumentList, error) {
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		documentsPath(orgId, projectId, database, collection),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if filter != "" {
		q.Set("filter", filter)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("list documents request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "list documents")
	}

	var result DocumentList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode documents response: %w", err)
	}
	return &result, nil
}

// GetDocument returns a document's chunks, optionally including vectors.
func (c *DeploymentClient) GetDocument(
	ctx context.Context,
	orgId, projectId, database, collection, documentID string,
	limit int,
	cursor string,
	includeVector bool,
) (*DocumentChunks, error) {
	path := documentsPath(orgId, projectId, database, collection) + "/" + url.PathEscape(documentID)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if includeVector {
		q.Set("include_vector", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("get document request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "get document")
	}

	var result DocumentChunks
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode document response: %w", err)
	}
	return &result, nil
}

// DeleteDocument deletes every chunk belonging to a document.
func (c *DeploymentClient) DeleteDocument(
	ctx context.Context,
	orgId, projectId, database, collection, documentID string,
) (*DeleteDocumentResult, error) {
	path := documentsPath(orgId, projectId, database, collection) + "/" + url.PathEscape(documentID)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("delete document request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "delete document")
	}

	var result DeleteDocumentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode delete response: %w", err)
	}
	return &result, nil
}
