package replay

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
	agent         *deployment.DescribeAgentResponse
	err           error
	describeCalls int
	replayCalls   int
}

func (f *fakeDeploy) ReplayEndpoint(
	orgID, projectID, agentName string,
) (string, func(*http.Request) error) {
	f.replayCalls++
	return "https://api.test/v1/organizations/" + orgID +
			"/projects/" + projectID + "/agents/" + agentName,
		func(req *http.Request) error {
			req.Header.Set("Authorization", "Bearer platform-token")
			return nil
		}
}

func (f *fakeDeploy) DescribeAgent(
	_ context.Context, _, _, _ string,
) (*deployment.DescribeAgentResponse, error) {
	f.describeCalls++
	return f.agent, f.err
}

type fakeAgent struct {
	baseURL, authHeader string
	follow              bool
	startReq            *agent.StartRequest
	startResp           *agent.StartResponse
	startErr            error
	waitRunID           string
	waitRun             *agent.Run
	waitErr             error
	progress            []*agent.Run
	retryErr            error
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
			},
		},
		agent: &fakeAgent{
			startResp: &agent.StartResponse{RunID: "run-1"},
			waitRun:   finishedRun(agent.StatusPassed),
		},
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
		NewAgent: func(baseURL string, auth agent.Auth, follow bool) AgentAPI {
			h.agent.baseURL, h.agent.follow = baseURL, follow
			probe := httptest.NewRequest(http.MethodGet, "/", nil)
			if err := auth(probe); err != nil {
				panic(err)
			}
			h.agent.authHeader = probe.Header.Get("Authorization")
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
	if h.agent.startReq.Dataset != "replay-chat" || h.agent.waitRunID != "run-1" {
		t.Errorf("start %+v wait %q", h.agent.startReq, h.agent.waitRunID)
	}
	if !strings.HasPrefix(
		h.stderr.String(),
		"agent-chat-dev  0.15.1  rev 592    via platform\n",
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
	if h.deploy.describeCalls != 0 || h.deploy.replayCalls != 0 {
		t.Errorf("network calls before file validation: describe=%d replay=%d",
			h.deploy.describeCalls, h.deploy.replayCalls)
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

func TestRunTarget(t *testing.T) {
	const platformURL = "https://api.test/v1/organizations/o/projects/p/agents/agent-chat-dev"

	tests := []struct {
		name       string
		endpoint   string
		agentURL   string
		apiKey     string
		wantURL    string
		wantAuth   string
		wantStderr string
		wantErr    string
	}{
		{
			name:       "default goes through the platform as the caller",
			endpoint:   "agent-chat-dev.example.com",
			wantURL:    platformURL,
			wantAuth:   "Bearer platform-token",
			wantStderr: "via platform",
		},
		{
			// An agent with no endpoint used to be unreplayable.
			name:       "an agent without an endpoint replays too",
			wantURL:    platformURL,
			wantAuth:   "Bearer platform-token",
			wantStderr: "via platform",
		},
		{
			name:       "agent-url talks straight to the agent with its own key",
			endpoint:   "agent-chat-dev.example.com",
			agentURL:   "http://127.0.0.1:8080",
			apiKey:     "local-key",
			wantURL:    "http://127.0.0.1:8080",
			wantAuth:   "Bearer local-key",
			wantStderr: "http://127.0.0.1:8080",
		},
		{
			name:     "agent-url without a key is refused before any call",
			agentURL: "http://127.0.0.1:8080",
			wantErr:  "--agent-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness()
			h.deploy.agent.Endpoint = tt.endpoint
			err := h.run(func(o *Options) {
				o.AgentURL, o.APIKey = tt.agentURL, tt.apiKey
			})

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to mention %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if h.agent.baseURL != tt.wantURL {
				t.Errorf("baseURL = %q, want %q", h.agent.baseURL, tt.wantURL)
			}
			if h.agent.authHeader != tt.wantAuth {
				t.Errorf("authHeader = %q, want %q", h.agent.authHeader, tt.wantAuth)
			}
			if !strings.Contains(h.stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to name %q", h.stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestRunStartErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		agentURL string
		want     []string
	}{
		{
			name: "401 through the platform blames the login",
			err:  &agent.Error{Status: 401, Detail: "Unauthorized"},
			want: []string{"not authorized to replay", "logged in"},
		},
		{
			name:     "401 on a direct call blames the flag",
			err:      &agent.Error{Status: 401, Detail: "Unauthorized"},
			agentURL: "http://127.0.0.1:8080",
			want:     []string{"the agent rejected --agent-api-key"},
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
			err := h.run(func(o *Options) {
				if tt.agentURL != "" {
					o.AgentURL, o.APIKey = tt.agentURL, "local-key"
				}
			})
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
		name       string
		err        error
		want       []string
		wantStderr []string
		wantNil    bool
	}{
		{
			name: "the stream closed with the run still going",
			err:  agent.ErrStreamEnded,
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
