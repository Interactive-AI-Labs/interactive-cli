package deployment

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// ReplayStartRequest is the body of POST .../agents/{name}/replays.
type ReplayStartRequest struct {
	Dataset      string         `json:"dataset,omitempty"`
	Scenarios    []string       `json:"scenarios,omitempty"`
	ScenarioBody map[string]any `json:"scenario_body,omitempty"`
	Repeat       int            `json:"repeat"`
	Concurrency  int            `json:"concurrency"`
}

// ReplaySkipped is a dataset item the agent did not replay, with the reason.
type ReplaySkipped struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

type ReplayStartResponse struct {
	RunID   string          `json:"run_id"`
	Skipped []ReplaySkipped `json:"skipped"`
}

// ReplayRun is a run as the agent reports it. A dataset run carries Batches; an
// inline run carries Iterations and is normalised into a suite of one batch.
type ReplayRun struct {
	RunID       string            `json:"run_id"`
	Dataset     string            `json:"dataset"`
	Scenario    string            `json:"scenario"`
	Status      string            `json:"status"`
	Repeat      int               `json:"repeat"`
	Concurrency int               `json:"concurrency"`
	Passed      int               `json:"passed"`
	Skipped     []ReplaySkipped   `json:"skipped"`
	Batches     []ReplayBatch     `json:"batches"`
	Iterations  []ReplayIteration `json:"iterations"`
	Error       string            `json:"error"`
	// Raw is the body exactly as the agent returned it, for --json.
	Raw []byte `json:"-"`
}

// ReplayBatch is one scenario replayed Repeat times.
type ReplayBatch struct {
	RunID      string            `json:"run_id"`
	Scenario   string            `json:"scenario"`
	Status     string            `json:"status"`
	Repeat     int               `json:"repeat"`
	Passed     int               `json:"passed"`
	Iterations []ReplayIteration `json:"iterations"`
	Error      string            `json:"error"`
}

// ReplayIteration is one replayed conversation and its verdict.
type ReplayIteration struct {
	RunID       string         `json:"run_id"`
	Scenario    string         `json:"scenario"`
	Status      string         `json:"status"`
	SessionKey  string         `json:"session_key"`
	TraceIDs    []string       `json:"trace_ids"`
	EvalTraceID string         `json:"eval_trace_id"`
	Turns       int            `json:"turns"`
	Observed    ReplayObserved `json:"observed"`
	Diverged    []string       `json:"diverged"`
	Judge       *ReplayJudge   `json:"judge"`
	Failures    []string       `json:"failures"`
	Error       string         `json:"error"`
}

type ReplayObserved struct {
	ToolsCalled []string `json:"tools_called"`
	ToolsDenied []string `json:"tools_denied"`
	Steps       []string `json:"steps"`
	Routines    []string `json:"routines"`
	Policies    []string `json:"policies"`
}

type ReplayJudge struct {
	Score     string `json:"score"`
	Reasoning string `json:"reasoning"`
}

const (
	ReplayStatusRunning = "running"
	ReplayStatusPassed  = "passed"
	ReplayStatusFailed  = "failed"
	ReplayStatusError   = "error"
)

// ReplayError is a refusal: the agent's own detail relayed by the platform, or the platform's message.
type ReplayError struct {
	Status  int
	Detail  string
	Skipped []ReplaySkipped
}

func (e *ReplayError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("agent returned %d", e.Status)
	}
	return fmt.Sprintf("agent returned %d: %s", e.Status, e.Detail)
}

// ErrReplayStreamEnded is a followed run whose stream closed before a verdict; following again picks it back up.
var ErrReplayStreamEnded = errors.New("status stream ended before the run finished")

func replayPath(orgID, projectID, agentName string) string {
	return fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/agents/%s/replays",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
		url.PathEscape(agentName),
	)
}

func (c *DeploymentClient) StartReplay(
	ctx context.Context,
	orgID, projectID, agentName string,
	req ReplayStartRequest,
) (*ReplayStartResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode replay request: %w", err)
	}
	reqHTTP, err := c.newRequest(ctx, http.MethodPost, replayPath(orgID, projectID, agentName))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Body = io.NopCloser(strings.NewReader(string(body)))

	resp, err := c.do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("replay request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the replay response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeReplayError(resp.StatusCode, raw)
	}

	var out ReplayStartResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("failed to decode replay response: %w", err)
	}
	if out.RunID == "" {
		return nil, fmt.Errorf("agent accepted the replay but returned no run_id")
	}
	return &out, nil
}

// replayEvent is one line of the platform's status stream.
type replayEvent struct {
	Run     json.RawMessage `json:"run"`
	Error   string          `json:"error"`
	Timeout bool            `json:"timeout"`
}

// FollowReplay reads the run's status stream to its verdict, calling onProgress
// on every update. There is no automatic reconnection, as with logs --follow.
func (c *DeploymentClient) FollowReplay(
	ctx context.Context,
	orgID, projectID, agentName, runID string,
	onProgress func(*ReplayRun),
) (*ReplayRun, error) {
	path := replayPath(orgID, projectID, agentName) + "/" + url.PathEscape(runID) + "?follow=true"
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("replay request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, decodeReplayError(resp.StatusCode, raw)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 256*1024), 8<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev replayEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, fmt.Errorf("failed to decode replay status: %w", err)
		}
		switch {
		case ev.Error != "":
			return nil, errors.New(ev.Error)
		case ev.Timeout:
			return nil, ErrReplayStreamEnded
		case len(ev.Run) > 0:
			run, err := decodeReplayRun(ev.Run)
			if err != nil {
				return nil, err
			}
			if onProgress != nil {
				onProgress(run)
			}
			if run.Status != ReplayStatusRunning {
				return run, nil
			}
		}
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, ErrReplayStreamEnded
}

// decodeReplayRun normalises the inline form into a suite holding one batch.
func decodeReplayRun(raw []byte) (*ReplayRun, error) {
	var run ReplayRun
	if err := json.Unmarshal(raw, &run); err != nil {
		return nil, fmt.Errorf("failed to decode replay run: %w", err)
	}
	run.Raw = raw
	if len(run.Batches) == 0 && len(run.Iterations) > 0 {
		run.Batches = []ReplayBatch{{
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

// decodeReplayError reads the agent's error bodies ({"detail": text | 422 list}) or the platform's ({"message": text}).
func decodeReplayError(status int, body []byte) *ReplayError {
	e := &ReplayError{Status: status}
	var env struct {
		Detail  json.RawMessage `json:"detail"`
		Skipped []ReplaySkipped `json:"skipped"`
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
	if msg := clients.ExtractServerMessage(body); msg != "" {
		e.Detail = msg
		return e
	}
	e.Detail = strings.TrimSpace(string(body))
	return e
}
