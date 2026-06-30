package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// CollectionSummary is one entry in a collections list response.
type CollectionSummary struct {
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type listCollectionsResponse struct {
	Collections []CollectionSummary `json:"collections"`
}

// CollectionSlot describes one vector slot in a collection's config.
type CollectionSlot struct {
	Type      string `json:"type"`
	Dimension int    `json:"dimension"`
	Distance  string `json:"distance"`
	Embedding *struct {
		Model string `json:"model"`
	} `json:"embedding"`
	Index *struct {
		Type string `json:"type"`
	} `json:"index"`
}

// CollectionFullText is the optional full-text search config.
type CollectionFullText struct {
	Enabled  bool   `json:"enabled"`
	Language string `json:"language"`
}

// CollectionConfig is the stored, normalized collection configuration.
type CollectionConfig struct {
	Vectors  map[string]CollectionSlot `json:"vectors"`
	FullText *CollectionFullText       `json:"full_text"`
}

// DescribeCollectionResponse is the full describe payload.
type DescribeCollectionResponse struct {
	Name      string           `json:"name"`
	Config    CollectionConfig `json:"config"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
}

// CollectionStats holds operational stats for a collection.
type CollectionStats struct {
	ChunkCount int64           `json:"chunkCount"`
	IndexValid map[string]bool `json:"indexValid"`
	SizeBytes  int64           `json:"sizeBytes"`
}

// collectionsPath builds the base collections path for a database.
func collectionsPath(orgId, projectId, database string) string {
	return fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/databases/%s/collections",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(database),
	)
}

// collectionErr turns a non-2xx collections response into an error, preferring
// the server's message.
func collectionErr(resp *http.Response, action string) error {
	respBody, _ := io.ReadAll(resp.Body)
	if msg := ExtractServerMessage(respBody); msg != "" {
		return errors.New(msg)
	}
	return fmt.Errorf("failed to %s: server returned %s", action, resp.Status)
}

// ListCollections returns the collections in a database.
func (c *DeploymentClient) ListCollections(
	ctx context.Context,
	orgId, projectId, database string,
) ([]CollectionSummary, error) {
	req, err := c.newRequest(ctx, http.MethodGet, collectionsPath(orgId, projectId, database))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("list collections request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "list collections")
	}

	var result listCollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode collections response: %w", err)
	}
	return result.Collections, nil
}

// DescribeCollection returns a collection's normalized config.
func (c *DeploymentClient) DescribeCollection(
	ctx context.Context,
	orgId, projectId, database, name string,
) (*DescribeCollectionResponse, error) {
	path := collectionsPath(orgId, projectId, database) + "/" + url.PathEscape(name)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("describe collection request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "describe collection")
	}

	var result DescribeCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode collection response: %w", err)
	}
	return &result, nil
}

// GetCollectionStats returns operational stats for a collection.
func (c *DeploymentClient) GetCollectionStats(
	ctx context.Context,
	orgId, projectId, database, name string,
) (*CollectionStats, error) {
	path := collectionsPath(orgId, projectId, database) + "/" + url.PathEscape(name) + "/stats"
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("collection stats request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "get collection stats")
	}

	var result CollectionStats
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}
	return &result, nil
}

// CreateCollection creates a collection from a raw JSON config body.
func (c *DeploymentClient) CreateCollection(
	ctx context.Context,
	orgId, projectId, database, name string,
	body []byte,
) (string, error) {
	path := collectionsPath(orgId, projectId, database) + "/" + url.PathEscape(name)
	return c.sendCollectionBody(ctx, http.MethodPost, path, body, "create collection")
}

// PatchCollection updates a collection's mutable config from a raw JSON body.
func (c *DeploymentClient) PatchCollection(
	ctx context.Context,
	orgId, projectId, database, name string,
	body []byte,
) (string, error) {
	path := collectionsPath(orgId, projectId, database) + "/" + url.PathEscape(name)
	return c.sendCollectionBody(ctx, http.MethodPatch, path, body, "update collection")
}

// DeleteCollection deletes a collection and its data.
func (c *DeploymentClient) DeleteCollection(
	ctx context.Context,
	orgId, projectId, database, name string,
) (string, error) {
	path := collectionsPath(orgId, projectId, database) + "/" + url.PathEscape(name)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("delete collection request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	msg := ExtractServerMessage(respBody)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg != "" {
			return "", errors.New(msg)
		}
		return "", fmt.Errorf("failed to delete collection: server returned %s", resp.Status)
	}
	return msg, nil
}

// sendCollectionBody issues a JSON-body request and returns the server message.
func (c *DeploymentClient) sendCollectionBody(
	ctx context.Context,
	method, path string,
	body []byte,
	action string,
) (string, error) {
	req, err := c.newRequest(ctx, method, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("%s request failed: %w", action, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	msg := ExtractServerMessage(respBody)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg != "" {
			return "", errors.New(msg)
		}
		return "", fmt.Errorf("failed to %s: server returned %s", action, resp.Status)
	}
	return msg, nil
}
