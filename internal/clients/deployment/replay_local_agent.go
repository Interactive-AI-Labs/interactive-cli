package deployment

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
)

// localAgentPoll is how often a local run is re-read; the agent serves no status stream.
const localAgentPoll = 2 * time.Second

// LocalAgentClient replays an agent on this machine with its own key; a deployed one goes through the platform.
type LocalAgentClient struct {
	baseURL    string
	key        string
	httpClient *http.Client
}

func NewLocalAgentClient(baseURL, key string) *LocalAgentClient {
	return &LocalAgentClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		key:     key,
		// No client-wide timeout: a run is minutes of real inference.
		httpClient: &http.Client{},
	}
}

// DescribeAgent reports what is knowable without the platform: no release, so no revision.
func (c *LocalAgentClient) DescribeAgent(
	_ context.Context,
	_, _, name string,
) (*DescribeAgentResponse, error) {
	return &DescribeAgentResponse{Name: name, Version: "local agent at " + c.baseURL}, nil
}

func (c *LocalAgentClient) StartReplay(
	ctx context.Context,
	_, _, _ string,
	req ReplayStartRequest,
) (*ReplayStartResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode replay request: %w", err)
	}
	raw, err := c.call(ctx, http.MethodPost, "/replays", body)
	if err != nil {
		return nil, err
	}
	return decodeReplayStart(raw)
}

// localRunPath is the agent's own status route; it carries no query to ask for a stream.
func localRunPath(runID string) string {
	return "/replays/" + url.PathEscape(runID)
}

// FollowReplay polls to a verdict; the status stream belongs to the platform, not the agent.
func (c *LocalAgentClient) FollowReplay(
	ctx context.Context,
	_, _, _, runID string,
	onProgress func(*ReplayRun),
) (*ReplayRun, error) {
	path := localRunPath(runID)
	for {
		raw, err := c.call(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}
		run, err := decodeReplayRun(raw)
		if err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(run)
		}
		if run.Status != ReplayStatusRunning {
			return run, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(localAgentPoll):
		}
	}
}

func (c *LocalAgentClient) call(
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
		return nil, decodeReplayError(resp.StatusCode, raw)
	}
	return raw, nil
}
