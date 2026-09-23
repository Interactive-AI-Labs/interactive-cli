package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

// pollEvery is how often a local run is re-read; the agent serves no status stream.
const pollEvery = 2 * time.Second

// Client replays an agent on this machine with its own key; a deployed one goes through the platform.
type Client struct {
	baseURL    string
	key        string
	httpClient *http.Client
}

// NewClient bounds each request, not the run: the agent answers both routes at once.
func NewClient(baseURL, key string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		key:        key,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) StartReplay(
	ctx context.Context,
	_, _, _ string,
	req deployment.ReplayStartRequest,
) (*deployment.ReplayStartResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode replay request: %w", err)
	}
	raw, err := c.call(ctx, http.MethodPost, "/replays", body)
	if err != nil {
		return nil, err
	}
	return deployment.DecodeReplayStart(raw)
}

// localRunPath is the agent's own status route; it carries no query to ask for a stream.
func localRunPath(runID string) string {
	return "/replays/" + url.PathEscape(runID)
}

// FollowReplay polls to a verdict; the status stream belongs to the platform, not the agent.
func (c *Client) FollowReplay(
	ctx context.Context,
	_, _, _, runID string,
	onProgress func(*deployment.ReplayRun),
) (*deployment.ReplayRun, error) {
	path := localRunPath(runID)
	for {
		raw, err := c.call(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}
		run, err := deployment.DecodeReplayRun(raw)
		if err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(run)
		}
		if run.Status != deployment.ReplayStatusRunning {
			return run, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollEvery):
		}
	}
}

func (c *Client) call(
	ctx context.Context,
	method, path string,
	body []byte,
) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// The agent's key is a bearer token, not the Basic --api-key the platform takes.
	if err := clients.ApplyRequestHeaders(req, c.key, "", nil); err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, req.Context().Err()
		}
		return nil, fmt.Errorf("replay request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the agent's response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, deployment.DecodeReplayError(resp.StatusCode, raw)
	}
	return raw, nil
}
