package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SlotAddResult is the response from adding a slot.
type SlotAddResult struct {
	Slot        string `json:"slot"`
	Type        string `json:"type"`
	Dimension   int    `json:"dimension"`
	Distance    string `json:"distance"`
	IndexStatus string `json:"indexStatus"`
}

// SlotIndexProgress is the response from index-progress.
type SlotIndexProgress struct {
	Slot      string `json:"slot"`
	IndexType string `json:"indexType"`
	Status    string `json:"status"`
}

// SlotOpResult is the response from reindex/vacuum.
type SlotOpResult struct {
	Slot        string `json:"slot"`
	IndexStatus string `json:"indexStatus,omitempty"`
	Status      string `json:"status,omitempty"`
}

func slotPath(orgId, projectId, database, collection, slot string) string {
	base := CollectionsPath(orgId, projectId, database)
	return base + "/" + url.PathEscape(collection) + "/vectors/" + url.PathEscape(slot)
}

// AddSlot adds a vector slot from a raw JSON body.
func (c *DeploymentClient) AddSlot(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
	body []byte,
) (*SlotAddResult, error) {
	var result SlotAddResult
	path := slotPath(orgId, projectId, database, collection, slot)
	if err := c.sendJSONInto(ctx, http.MethodPut, path, body, "add slot", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReindexSlot rebuilds a slot's index from an optional config body.
func (c *DeploymentClient) ReindexSlot(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
	body []byte,
) (*SlotOpResult, error) {
	var result SlotOpResult
	path := slotPath(orgId, projectId, database, collection, slot) + "/reindex"
	if err := c.sendJSONInto(ctx, http.MethodPost, path, body, "reindex slot", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VacuumSlot reclaims space and refreshes stats for a slot. The server
// accepts a bodyless POST to /vacuum; no body is sent.
func (c *DeploymentClient) VacuumSlot(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
) (*SlotOpResult, error) {
	var result SlotOpResult
	path := slotPath(orgId, projectId, database, collection, slot) + "/vacuum"
	if err := c.sendJSONInto(ctx, http.MethodPost, path, nil, "vacuum slot", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SlotIndexProgressStatus reads a slot's index build progress.
func (c *DeploymentClient) SlotIndexProgressStatus(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
) (*SlotIndexProgress, error) {
	path := slotPath(orgId, projectId, database, collection, slot) + "/index-progress"
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("index progress request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, collectionErr(resp, "index progress")
	}

	var result SlotIndexProgress
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode index progress response: %w", err)
	}
	return &result, nil
}

// DeleteSlot drops a vector slot (its column and index).
func (c *DeploymentClient) DeleteSlot(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
) (string, error) {
	path := slotPath(orgId, projectId, database, collection, slot)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("delete slot request failed: %w", err)
	}
	defer resp.Body.Close()

	return serverMessage(resp, "delete slot")
}
