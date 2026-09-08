// Package agent is the HTTP client for the replay API served by an agent
// itself, as opposed to the platform or deployment APIs.
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/buildinfo"
)

type Client struct {
	baseURL string
	bearer  string
	http    *http.Client
}

func NewClient(baseURL, bearer string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		bearer:  bearer,
		http:    &http.Client{Timeout: timeout},
	}
}

// Error is a non-2xx response from the agent. Detail carries the agent's own
// wording (FastAPI flat detail, or the joined messages of a 422 list); Skipped
// is set on the 400s that report unreplayable dataset items.
type Error struct {
	Status  int
	Detail  string
	Skipped []Skipped
}

func (e *Error) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("agent returned %d", e.Status)
	}
	return fmt.Sprintf("agent returned %d: %s", e.Status, e.Detail)
}

func (c *Client) StartReplay(ctx context.Context, req StartRequest) (*StartResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode replay request: %w", err)
	}
	raw, err := c.do(ctx, http.MethodPost, "/replays", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var resp StartResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode replay response: %w", err)
	}
	if resp.RunID == "" {
		return nil, fmt.Errorf("agent accepted the replay but returned no run_id")
	}
	return &resp, nil
}

func (c *Client) GetReplay(ctx context.Context, runID string) (*Run, error) {
	raw, err := c.do(ctx, http.MethodGet, "/replays/"+runID, nil)
	if err != nil {
		return nil, err
	}
	var run Run
	if err := json.Unmarshal(raw, &run); err != nil {
		return nil, fmt.Errorf("failed to decode replay run: %w", err)
	}
	run.Raw = raw
	if len(run.Batches) == 0 && len(run.Iterations) > 0 {
		run.Batches = []Batch{{
			RunID:      run.RunID,
			Scenario:   run.Scenario,
			Status:     run.Status,
			Repeat:     run.Repeat,
			Passed:     run.Passed,
			Iterations: run.Iterations,
			Error:      run.Error,
		}}
		run.Iterations = nil
	}
	return &run, nil
}

type WaitOptions struct {
	Interval time.Duration
	Timeout  time.Duration
	// NotFoundCap bounds a streak of 404/5xx polls. The agent's run registry is
	// in memory: a 404 can be a replica that never saw the POST, or a restart
	// that lost the run for good. Retry, but not forever.
	NotFoundCap time.Duration
	OnProgress  func(*Run)
}

// Wait polls GET /replays/{run_id} until the run is no longer running.
// Returns context.DeadlineExceeded when Timeout elapses and context.Canceled
// when the parent context is cancelled; the run keeps going on the agent.
func (c *Client) Wait(ctx context.Context, runID string, opts WaitOptions) (*Run, error) {
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	var firstFailure time.Time
	for {
		run, err := c.GetReplay(ctx, runID)
		switch {
		case err == nil:
			firstFailure = time.Time{}
			if opts.OnProgress != nil {
				opts.OnProgress(run)
			}
			if run.Status != StatusRunning {
				return run, nil
			}
		case ctx.Err() != nil:
			return nil, ctx.Err()
		case isTransient(err):
			if firstFailure.IsZero() {
				firstFailure = time.Now()
			}
			if time.Since(firstFailure) >= opts.NotFoundCap {
				return nil, fmt.Errorf(
					"run %s not found for %s; the agent probably restarted and lost it "+
						"(the run registry is in memory) — scores already written are still on the platform: %w",
					runID, opts.NotFoundCap, err,
				)
			}
		default:
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(opts.Interval):
		}
	}
}

func isTransient(err error) bool {
	var ae *Error
	if !errors.As(err, &ae) {
		return false
	}
	return ae.Status == http.StatusNotFound || ae.Status >= 500
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", buildinfo.UserAgent)
	req.Header.Set("Authorization", "Bearer "+c.bearer)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeError(resp.StatusCode, raw)
	}
	return raw, nil
}

// decodeError reads FastAPI error bodies: {"detail": "text"} or, for 422,
// {"detail": [{"loc": [...], "msg": "..."}]}; anything else is kept verbatim.
func decodeError(status int, body []byte) *Error {
	e := &Error{Status: status}
	var env struct {
		Detail  json.RawMessage `json:"detail"`
		Skipped []Skipped       `json:"skipped"`
	}
	if json.Unmarshal(body, &env) == nil && len(env.Detail) > 0 {
		e.Skipped = env.Skipped
		var text string
		if json.Unmarshal(env.Detail, &text) == nil {
			e.Detail = text
			return e
		}
		var items []struct {
			Loc []any  `json:"loc"`
			Msg string `json:"msg"`
		}
		if json.Unmarshal(env.Detail, &items) == nil {
			msgs := make([]string, 0, len(items))
			for _, it := range items {
				if n := len(it.Loc); n > 0 {
					msgs = append(msgs, fmt.Sprintf("%v: %s", it.Loc[n-1], it.Msg))
				} else {
					msgs = append(msgs, it.Msg)
				}
			}
			e.Detail = strings.Join(msgs, "; ")
			return e
		}
	}
	e.Detail = strings.TrimSpace(string(body))
	return e
}
