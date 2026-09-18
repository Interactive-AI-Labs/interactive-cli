package replay

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestStartError(t *testing.T) {
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
			got := startError(tt.err, "agent-chat-dev", "0.15.1", tt.agentURL)
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

// stoppedClient reports the watch as interrupted, the path that prints how to resume.
type stoppedClient struct{}

func (stoppedClient) DescribeAgent(
	context.Context, string, string, string,
) (*deployment.DescribeAgentResponse, error) {
	return &deployment.DescribeAgentResponse{Version: "0.15.2", Revision: 42}, nil
}

func (stoppedClient) StartReplay(
	context.Context, string, string, string, deployment.ReplayStartRequest,
) (*deployment.ReplayStartResponse, error) {
	return &deployment.ReplayStartResponse{RunID: "run-1"}, nil
}

func (stoppedClient) FollowReplay(
	context.Context, string, string, string, string, func(*deployment.ReplayRun),
) (*deployment.ReplayRun, error) {
	return nil, context.Canceled
}

// The suggested command has to reach the same agent the run is on.
func TestReattachCommandKeepsTheTarget(t *testing.T) {
	tests := []struct {
		name     string
		agentURL string
		want     string
	}{
		{
			name: "through the platform",
			want: "iai agents replay my-agent --run-id run-1\n",
		},
		{
			name:     "straight to a local agent",
			agentURL: "http://127.0.0.1:8080",
			want:     "iai agents replay my-agent --run-id run-1 --agent-url http://127.0.0.1:8080\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			err := Run(
				context.Background(),
				Deps{Deploy: stoppedClient{}, Stdout: io.Discard, Stderr: &stderr},
				Options{
					AgentName: "my-agent",
					AgentURL:  tt.agentURL,
					Timeout:   time.Minute,
				},
			)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if !strings.HasSuffix(stderr.String(), tt.want) {
				t.Errorf("stderr ended %q, want it to end %q", stderr.String(), tt.want)
			}
		})
	}
}
