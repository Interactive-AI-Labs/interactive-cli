package replay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestRefusalError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		agentURL string
		want     []string
	}{
		{
			name: "401 blames the login",
			err:  &deployment.ReplayError{Status: 401, Detail: "Unauthorized"},
			want: []string{"not authorized to replay agent-chat-dev", "logged in"},
		},
		{
			name:     "401 on a local agent blames the key, not the login",
			err:      &deployment.ReplayError{Status: 401, Detail: "Unauthorized"},
			agentURL: "http://127.0.0.1:8080",
			want:     []string{"the agent at http://127.0.0.1:8080 rejected --agent-api-key"},
		},
		{
			name: "404 names the agent version",
			err:  &deployment.ReplayError{Status: 404, Detail: "Not Found"},
			want: []string{
				"agent agent-chat-dev (0.15.1) does not support replay",
				"0.15.0 or later",
			},
		},
		{
			name: "400 keeps the agent's wording and lists skipped items",
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
			name: "503 passes through verbatim",
			err: &deployment.ReplayError{
				Status: 503,
				Detail: "Replay is not available — missing: platform client.",
			},
			want: []string{
				"agent returned 503: Replay is not available — missing: platform client.",
			},
		},
		{
			name: "a transport error passes through",
			err:  errors.New("dial tcp: connection refused"),
			want: []string{"connection refused"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refusalError(tt.err, "agent-chat-dev", "0.15.1", tt.agentURL)
			for _, w := range tt.want {
				if !strings.Contains(got.Error(), w) {
					t.Errorf("error %q missing %q", got.Error(), w)
				}
			}
		})
	}
}

func TestErrorSummary(t *testing.T) {
	tests := []struct {
		name string
		run  *deployment.ReplayRun
		want string
	}{
		{
			name: "run-level error wins",
			run: &deployment.ReplayRun{
				Error:   "scenario is invalid",
				Batches: []deployment.ReplayBatch{{Error: "b"}},
			},
			want: "scenario is invalid",
		},
		{
			name: "then the first batch error",
			run: &deployment.ReplayRun{
				Batches: []deployment.ReplayBatch{{}, {Error: "batch failed"}},
			},
			want: "batch failed",
		},
		{
			name: "then the first iteration error",
			run: &deployment.ReplayRun{Batches: []deployment.ReplayBatch{{
				Iterations: []deployment.ReplayIteration{{}, {Error: "router timeout"}},
			}}},
			want: "router timeout",
		},
		{
			name: "nothing explicit points at the rendered output",
			run: &deployment.ReplayRun{
				Batches: []deployment.ReplayBatch{{Iterations: []deployment.ReplayIteration{{}}}},
			},
			want: "see the errored iterations above",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorSummary(tt.run); got != tt.want {
				t.Errorf("errorSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

// The suggested command has to reach the same agent the run is on.
func TestReattachCommand(t *testing.T) {
	tests := []struct {
		name     string
		agentURL string
		want     string
	}{
		{
			name: "through the platform",
			want: "iai agents replay my-agent --run-id run-1",
		},
		{
			name:     "straight to a local agent",
			agentURL: "http://127.0.0.1:8080",
			want:     "iai agents replay my-agent --run-id run-1 --agent-url http://127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reattachCommand("my-agent", "run-1", tt.agentURL)
			if got != tt.want {
				t.Errorf("command = %q, want %q", got, tt.want)
			}
		})
	}
}

// A refusal is the agent saying no; only a lost watch is worth following again.
func TestWaitError(t *testing.T) {
	tests := []struct {
		name     string
		waitErr  error
		agentURL string
		wantNil  bool
		wantErr  string
		wantHint bool
	}{
		{
			name:    "a stopped watch is not a failure",
			waitErr: context.Canceled,
			wantNil: true,
		},
		{
			name:     "giving up says how to resume",
			waitErr:  context.DeadlineExceeded,
			wantErr:  "gave up waiting",
			wantHint: true,
		},
		{
			name:     "a lost connection may leave the run going",
			waitErr:  errors.New("connection reset"),
			wantErr:  "connection reset",
			wantHint: true,
		},
		{
			name:     "a stream that ended may leave the run going",
			waitErr:  deployment.ErrReplayStreamEnded,
			wantErr:  "stream ended",
			wantHint: true,
		},
		{
			name:     "a gateway failure in front of the agent may leave the run going",
			waitErr:  &deployment.ReplayError{Status: 503, Detail: "Service Unavailable"},
			wantErr:  "Service Unavailable",
			wantHint: true,
		},
		{
			name:    "a run the agent does not know is not worth following",
			waitErr: &deployment.ReplayError{Status: 404, Detail: "Run not found"},
			wantErr: "Run not found",
		},
		{
			name:     "a rejected key is not worth following, and names the flag",
			waitErr:  &deployment.ReplayError{Status: 401, Detail: "Unauthorized"},
			agentURL: "http://127.0.0.1:8080",
			wantErr:  "rejected --agent-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				AgentName: "my-agent",
				AgentURL:  tt.agentURL,
				Timeout:   time.Minute,
			}
			got := waitError(tt.waitErr, opts, "0.16.0", "run-1")

			if tt.wantNil {
				if got != nil {
					t.Fatalf("error = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("error = nil, want the failure reported")
			}
			if !strings.Contains(got.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to keep %q", got, tt.wantErr)
			}
			if has := strings.Contains(got.Error(), "follow it again"); has != tt.wantHint {
				t.Errorf("hint present = %v, want %v (error: %v)", has, tt.wantHint, got)
			}
		})
	}
}
