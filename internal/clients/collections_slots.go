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
	base := collectionsPath(orgId, projectId, database)
	return base + "/" + url.PathEscape(collection) + "/vectors/" + url.PathEscape(slot)
}

// sendSlotBody issues a method+body request to a slot path and decodes dst.
func (c *DeploymentClient) sendSlotBody(
	ctx context.Context,
	method, path string,
	body []byte,
	action string,
	dst any,
) error {
	req, err := c.newRequest(ctx, method, path)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

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

// AddSlot adds a vector slot from a raw JSON body.
func (c *DeploymentClient) AddSlot(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
	body []byte,
) (*SlotAddResult, error) {
	var result SlotAddResult
	path := slotPath(orgId, projectId, database, collection, slot)
	if err := c.sendSlotBody(ctx, http.MethodPut, path, body, "add slot", &result); err != nil {
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
	if err := c.sendSlotBody(ctx, http.MethodPost, path, body, "reindex slot", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VacuumSlot reclaims space and refreshes stats for a slot.
func (c *DeploymentClient) VacuumSlot(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
) (*SlotOpResult, error) {
	var result SlotOpResult
	path := slotPath(orgId, projectId, database, collection, slot) + "/vacuum"
	if err := c.sendSlotBody(ctx, http.MethodPost, path, nil, "vacuum slot", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SlotIndexProgressStatus reads a slot's index build progress.
func (c *DeploymentClient) SlotIndexProgressStatus(
	ctx context.Context,
	orgId, projectId, database, collection, slot string,
) (*SlotIndexProgress, error) {
	var result SlotIndexProgress
	path := slotPath(orgId, projectId, database, collection, slot) + "/index-progress"
	if err := c.sendSlotBody(ctx, http.MethodGet, path, nil, "index progress", &result); err != nil {
		return nil, err
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

	respBody, _ := io.ReadAll(resp.Body)
	msg := ExtractServerMessage(respBody)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", fmt.Errorf("failed to delete slot: server returned %s", resp.Status)
	}
	return msg, nil
}
