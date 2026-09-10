package replay

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

type fakeDeploy struct {
	agent         *deployment.DescribeAgentResponse
	err           error
	describeCalls int
	startReq      *deployment.ReplayStartRequest
	startResp     *deployment.ReplayStartResponse
	startErr      error
	waitRunID     string
	waitRun       *deployment.ReplayRun
	waitErr       error
	progress      []*deployment.ReplayRun
}

func (f *fakeDeploy) DescribeAgent(
	_ context.Context, _, _, _ string,
) (*deployment.DescribeAgentResponse, error) {
	f.describeCalls++
	return f.agent, f.err
}

func (f *fakeDeploy) StartReplay(
	_ context.Context, _, _, _ string, req deployment.ReplayStartRequest,
) (*deployment.ReplayStartResponse, error) {
	f.startReq = &req
	return f.startResp, f.startErr
}

func (f *fakeDeploy) FollowReplay(
	_ context.Context, _, _, _, runID string, onProgress func(*deployment.ReplayRun),
) (*deployment.ReplayRun, error) {
	f.waitRunID = runID
	for _, r := range f.progress {
		onProgress(r)
	}
	return f.waitRun, f.waitErr
}

func finishedRun(status string) *deployment.ReplayRun {
	return &deployment.ReplayRun{
		RunID: "run-1", Dataset: "replay-chat", Status: status, Repeat: 1, Concurrency: 8,
		Batches: []deployment.ReplayBatch{{
			Scenario: "account-lock", Status: status, Repeat: 1, Passed: 1,
			Iterations: []deployment.ReplayIteration{{Status: status, EvalTraceID: "eval-1"}},
		}},
		Raw: []byte(`{"run_id":"run-1","status":"` + status + `"}`),
	}
}

type harness struct {
	deploy *fakeDeploy
	// agent is the same fake seen from the replay side, for readable assertions.
	agent  *fakeDeploy
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func newHarness() *harness {
	d := &fakeDeploy{
		agent:     &deployment.DescribeAgentResponse{Version: "0.15.1", Revision: 592},
		startResp: &deployment.ReplayStartResponse{RunID: "run-1"},
		waitRun:   finishedRun(deployment.ReplayStatusPassed),
	}
	return &harness{deploy: d, agent: d}
}

func (h *harness) run(mutate func(*Options)) error {
	opts := Options{
		OrgID: "o", ProjectID: "p", AgentName: "agent-chat-dev",
		Input:   inputs.ReplayInput{Dataset: "replay-chat", Repeat: 1, Concurrency: 8},
		Timeout: time.Minute,
	}
	if mutate != nil {
		mutate(&opts)
	}
	deps := Deps{Deploy: h.deploy, Stdout: &h.stdout, Stderr: &h.stderr}
	return Run(context.Background(), deps, opts)
}

func TestRunDatasetPassed(t *testing.T) {
	h := newHarness()
	if err := h.run(nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if h.agent.startReq.Dataset != "replay-chat" || h.agent.waitRunID != "run-1" {
		t.Errorf("start %+v wait %q", h.agent.startReq, h.agent.waitRunID)
	}
	if !strings.HasPrefix(
		h.stderr.String(),
		"agent-chat-dev  0.15.1  rev 592\n",
	) {
		t.Errorf("stderr = %q", h.stderr.String())
	}
	if !strings.Contains(h.stdout.String(), "PASS   1/1 passed     run run-1") {
		t.Errorf("stdout = %q", h.stdout.String())
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "platform-token") {
		t.Error("credential leaked into output")
	}
}

func TestRunFailedVerdictIsNotAnError(t *testing.T) {
	h := newHarness()
	h.agent.waitRun = finishedRun(deployment.ReplayStatusFailed)
	if err := h.run(nil); err != nil {
		t.Fatalf("Run() error = %v, want nil for a failed verdict", err)
	}
	if !strings.Contains(h.stdout.String(), "FAIL   1/1 passed") {
		t.Errorf("stdout = %q", h.stdout.String())
	}
}

func TestRunErrorStatusIsAnError(t *testing.T) {
	h := newHarness()
	h.agent.waitRun = finishedRun(deployment.ReplayStatusError)
	h.agent.waitRun.Batches[0].Iterations[0].Error = "JudgeError: no verdict"
	err := h.run(nil)
	if err == nil || !strings.Contains(err.Error(), "JudgeError: no verdict") {
		t.Errorf("error = %v", err)
	}
	if !strings.Contains(h.stdout.String(), "ERROR          JudgeError: no verdict") {
		t.Errorf("payload not rendered before the error: %q", h.stdout.String())
	}
}

func TestRunJSONWritesRawOnly(t *testing.T) {
	h := newHarness()
	h.agent.progress = []*deployment.ReplayRun{
		{
			Status:  deployment.ReplayStatusRunning,
			Batches: []deployment.ReplayBatch{{Status: deployment.ReplayStatusRunning}},
		},
	}
	if err := h.run(func(o *Options) { o.JSON = true }); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := h.stdout.String(); got != `{"run_id":"run-1","status":"passed"}`+"\n" {
		t.Errorf("stdout = %q", got)
	}
	if !strings.Contains(h.stderr.String(), "0/1 scenarios finished\n") {
		t.Errorf("progress missing from stderr: %q", h.stderr.String())
	}
	if !strings.Contains(h.stderr.String(), "eval trace (account-lock): iai traces get eval-1") {
		t.Errorf("pointer missing from stderr in --json mode: %q", h.stderr.String())
	}
}

func TestRunFileErrorBeforeAnyNetworkCall(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.yaml")
	if err := os.WriteFile(path, []byte("- a\n- b\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := newHarness()
	err := h.run(func(o *Options) {
		o.Input = inputs.ReplayInput{File: path, Repeat: 1, Concurrency: 8}
	})
	if err == nil || !strings.Contains(err.Error(), "must be a YAML or JSON object") {
		t.Fatalf("error = %v", err)
	}
	if h.deploy.describeCalls != 0 || h.deploy.startReq != nil {
		t.Errorf("network calls before file validation: describe=%d started=%v",
			h.deploy.describeCalls, h.deploy.startReq != nil)
	}
	if h.stderr.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", h.stderr.String())
	}
}

func TestRunFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "account-lock.yaml")
	if err := os.WriteFile(
		path,
		[]byte("customer_id: replay-1\nmessages: [hi]\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	h := newHarness()
	if err := h.run(
		func(o *Options) { o.Input = inputs.ReplayInput{File: path, Repeat: 2, Concurrency: 4} },
	); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	req := h.agent.startReq
	if req.Dataset != "" || req.ScenarioBody["scenario"] != "account-lock" || req.Repeat != 2 {
		t.Errorf("start request = %+v", req)
	}
}

func TestRunReattachSkipsStart(t *testing.T) {
	h := newHarness()
	if err := h.run(
		func(o *Options) { o.Input = inputs.ReplayInput{RunID: "old-run", Repeat: 1, Concurrency: 8} },
	); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if h.agent.startReq != nil || h.agent.waitRunID != "old-run" {
		t.Errorf("start %+v wait %q", h.agent.startReq, h.agent.waitRunID)
	}
}

func TestRunStartErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "401 blames the login",
			err:  &deployment.ReplayError{Status: 401, Detail: "Unauthorized"},
			want: []string{"not authorized to replay", "logged in"},
		},
		{
			name: "404",
			err:  &deployment.ReplayError{Status: 404, Detail: "Not Found"},
			want: []string{
				"agent agent-chat-dev (0.15.1) does not support replay",
				"0.15.0 or later",
			},
		},
		{
			name: "400 with skipped",
			err: &deployment.ReplayError{
				Status: 400,
				Detail: "dataset 'x' holds no replayable scenario",
				Skipped: []deployment.ReplaySkipped{
					{ID: "probe-1", Reason: "messages: Field required"},
				},
			},
			want: []string{
				"holds no replayable scenario",
				"\n  skipped probe-1: messages: Field required",
			},
		},
		{
			name: "503 verbatim",
			err: &deployment.ReplayError{
				Status: 503,
				Detail: "Replay is not available — missing: platform client.",
			},
			want: []string{
				"agent returned 503: Replay is not available — missing: platform client.",
			},
		},
		{
			name: "transport error passes through",
			err:  errors.New("dial tcp: connection refused"),
			want: []string{"connection refused"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness()
			h.agent.startErr = tt.err
			err := h.run(nil)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q missing %q", err.Error(), w)
				}
			}
		})
	}
}

func TestRunSkippedPrintedOnStart(t *testing.T) {
	h := newHarness()
	h.agent.startResp.Skipped = []deployment.ReplaySkipped{
		{ID: "account-lock", Reason: "not active"},
		{ID: "other", Reason: "invalid"},
	}
	err := h.run(func(o *Options) { o.Input.Scenarios = []string{"account-lock"} })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stderr.String(), "skipped account-lock: not active\n") ||
		strings.Contains(h.stderr.String(), "other") {
		t.Errorf("stderr = %q", h.stderr.String())
	}
}

func TestRunWaitErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		want       []string
		wantStderr []string
		wantNil    bool
	}{
		{
			name: "the stream closed with the run still going",
			err:  deployment.ErrReplayStreamEnded,
			want: []string{
				"run run-1 may still be running",
				"iai agents replay agent-chat-dev --run-id run-1",
			},
		},
		{
			name: "timeout",
			err:  context.DeadlineExceeded,
			want: []string{"gave up waiting after 1m0s", "--run-id run-1"},
		},
		{
			name: "lost run passes through",
			err:  errors.New("run run-1 not found for 5m0s"),
			want: []string{"not found for 5m0s"},
		},
		{
			// Ctrl-C stops watching without failing the command, as logs --follow does.
			name:    "interrupted is quiet",
			err:     context.Canceled,
			wantNil: true,
			wantStderr: []string{
				"stopped watching run run-1",
				"iai agents replay agent-chat-dev --run-id run-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness()
			h.agent.waitErr = tt.err
			err := h.run(nil)
			if tt.wantNil && err != nil {
				t.Fatalf("Run() error = %v, want nil", err)
			}
			if !tt.wantNil && err == nil {
				t.Fatal("expected error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q missing %q", err.Error(), w)
				}
			}
			for _, w := range tt.wantStderr {
				if !strings.Contains(h.stderr.String(), w) {
					t.Errorf("stderr %q missing %q", h.stderr.String(), w)
				}
			}
			if h.stdout.Len() != 0 {
				t.Errorf("stdout should be empty, got %q", h.stdout.String())
			}
		})
	}
}

func TestRunDescribeFails(t *testing.T) {
	h := newHarness()
	h.deploy.err = errors.New("agent not found")
	err := h.run(nil)
	if err == nil ||
		!strings.Contains(
			err.Error(),
			`failed to describe agent "agent-chat-dev": agent not found`,
		) {
		t.Errorf("error = %v", err)
	}
}
