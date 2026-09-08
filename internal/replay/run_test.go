package replay

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/agent"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

type fakeDeploy struct {
	fakeSecrets
	agent         *deployment.DescribeAgentResponse
	err           error
	describeCalls int
}

func (f *fakeDeploy) DescribeAgent(
	_ context.Context, _, _, _ string,
) (*deployment.DescribeAgentResponse, error) {
	f.describeCalls++
	return f.agent, f.err
}

type fakeAgent struct {
	baseURL, bearer string
	startReq        *agent.StartRequest
	startResp       *agent.StartResponse
	startErr        error
	waitRunID       string
	waitRun         *agent.Run
	waitErr         error
	progress        []*agent.Run
	retryErr        error
}

func (f *fakeAgent) StartReplay(
	_ context.Context,
	req agent.StartRequest,
) (*agent.StartResponse, error) {
	f.startReq = &req
	return f.startResp, f.startErr
}

func (f *fakeAgent) Wait(
	_ context.Context,
	runID string,
	opts agent.WaitOptions,
) (*agent.Run, error) {
	f.waitRunID = runID
	if f.retryErr != nil {
		opts.OnRetry(f.retryErr)
	}
	for _, r := range f.progress {
		opts.OnProgress(r)
	}
	return f.waitRun, f.waitErr
}

func finishedRun(status string) *agent.Run {
	return &agent.Run{
		RunID: "run-1", Dataset: "replay-chat", Status: status, Repeat: 1, Concurrency: 8,
		Batches: []agent.Batch{{
			Scenario: "account-lock", Status: status, Repeat: 1, Passed: 1,
			Iterations: []agent.Iteration{{Status: status, EvalTraceID: "eval-1"}},
		}},
		Raw: []byte(`{"run_id":"run-1","status":"` + status + `"}`),
	}
}

type harness struct {
	deploy *fakeDeploy
	agent  *fakeAgent
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func newHarness() *harness {
	h := &harness{
		deploy: &fakeDeploy{
			agent: &deployment.DescribeAgentResponse{
				Version:  "0.15.1",
				Revision: 592,
				Endpoint: "agent-chat-dev.example.com",
				AgentConfig: map[string]any{
					"runtime": map[string]any{"api_key": "${AGENT_API_KEY}"},
				},
				SecretRefs: []deployment.SecretRef{{SecretName: "platform-dev"}},
			},
		},
		agent: &fakeAgent{
			startResp: &agent.StartResponse{RunID: "run-1"},
			waitRun:   finishedRun(agent.StatusPassed),
		},
	}
	h.deploy.secrets = map[string]map[string]string{
		"platform-dev": {"AGENT_API_KEY": base64.StdEncoding.EncodeToString([]byte("secret-key"))},
	}
	return h
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
	deps := Deps{
		Deploy: h.deploy,
		NewAgent: func(baseURL, bearer string) AgentAPI {
			h.agent.baseURL, h.agent.bearer = baseURL, bearer
			return h.agent
		},
		Stdout: &h.stdout,
		Stderr: &h.stderr,
	}
	return Run(context.Background(), deps, opts)
}

func TestRunDatasetPassed(t *testing.T) {
	h := newHarness()
	if err := h.run(nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if h.agent.baseURL != "https://agent-chat-dev.example.com" || h.agent.bearer != "secret-key" {
		t.Errorf("agent built with %q / %q", h.agent.baseURL, h.agent.bearer)
	}
	if h.agent.startReq.Dataset != "replay-chat" || h.agent.waitRunID != "run-1" {
		t.Errorf("start %+v wait %q", h.agent.startReq, h.agent.waitRunID)
	}
	if !strings.HasPrefix(
		h.stderr.String(),
		"agent-chat-dev  0.15.1  rev 592    https://agent-chat-dev.example.com\n",
	) {
		t.Errorf("stderr = %q", h.stderr.String())
	}
	if !strings.Contains(h.stdout.String(), "PASS   1/1 passed     run run-1") {
		t.Errorf("stdout = %q", h.stdout.String())
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "secret-key") {
		t.Error("bearer leaked into output")
	}
}

func TestRunFailedVerdictIsNotAnError(t *testing.T) {
	h := newHarness()
	h.agent.waitRun = finishedRun(agent.StatusFailed)
	if err := h.run(nil); err != nil {
		t.Fatalf("Run() error = %v, want nil for a failed verdict", err)
	}
	if !strings.Contains(h.stdout.String(), "FAIL   1/1 passed") {
		t.Errorf("stdout = %q", h.stdout.String())
	}
}

func TestRunErrorStatusIsAnError(t *testing.T) {
	h := newHarness()
	h.agent.waitRun = finishedRun(agent.StatusError)
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
	h.agent.progress = []*agent.Run{
		{Status: agent.StatusRunning, Batches: []agent.Batch{{Status: agent.StatusRunning}}},
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
	if h.deploy.describeCalls != 0 || len(h.deploy.fetched) != 0 {
		t.Errorf("network calls before file validation: describe=%d secrets=%v",
			h.deploy.describeCalls, h.deploy.fetched)
	}
	if h.stderr.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", h.stderr.String())
	}
}

func TestRunPrintsRetryNote(t *testing.T) {
	h := newHarness()
	h.agent.retryErr = &agent.Error{Status: 404, Detail: "No such replay run: run-1"}
	if err := h.run(nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := "agent returned 404: No such replay run: run-1; retrying for up to 5m0s\n"
	if !strings.Contains(h.stderr.String(), want) {
		t.Errorf("stderr = %q, want to contain %q", h.stderr.String(), want)
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

func TestRunAgentURLOverridesEndpoint(t *testing.T) {
	h := newHarness()
	h.deploy.agent.Endpoint = ""
	if err := h.run(func(o *Options) { o.AgentURL = "http://127.0.0.1:8080" }); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if h.agent.baseURL != "http://127.0.0.1:8080" {
		t.Errorf("baseURL = %q", h.agent.baseURL)
	}
}

func TestRunNoEndpoint(t *testing.T) {
	h := newHarness()
	h.deploy.agent.Endpoint = ""
	err := h.run(nil)
	if err == nil || !strings.Contains(err.Error(), "has no public endpoint") ||
		!strings.Contains(err.Error(), "--agent-url http://127.0.0.1:8080") {
		t.Errorf("error = %v", err)
	}
}

func TestRunStartErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "401",
			err:  &agent.Error{Status: 401, Detail: "Unauthorized"},
			want: []string{"the agent rejected the api key", "--agent-api-key"},
		},
		{
			name: "404",
			err:  &agent.Error{Status: 404, Detail: "Not Found"},
			want: []string{
				"agent agent-chat-dev (0.15.1) has no /replays route",
				"0.15.0 or later",
			},
		},
		{
			name: "400 with skipped",
			err: &agent.Error{
				Status: 400, Detail: "dataset 'x' holds no replayable scenario",
				Skipped: []agent.Skipped{{ID: "probe-1", Reason: "messages: Field required"}},
			},
			want: []string{
				"holds no replayable scenario",
				"\n  skipped probe-1: messages: Field required",
			},
		},
		{
			name: "503 verbatim",
			err: &agent.Error{
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
	h.agent.startResp.Skipped = []agent.Skipped{
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
		name string
		err  error
		want []string
	}{
		{
			name: "interrupted",
			err:  context.Canceled,
			want: []string{"interrupted", "iai agents replay agent-chat-dev --run-id run-1"},
		},
		{
			name: "timeout",
			err:  context.DeadlineExceeded,
			want: []string{"gave up polling after 1m0s", "--run-id run-1"},
		},
		{
			name: "lost run passes through",
			err:  errors.New("run run-1 not found for 5m0s"),
			want: []string{"not found for 5m0s"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness()
			h.agent.waitErr = tt.err
			err := h.run(nil)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q missing %q", err.Error(), w)
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
