package replay

import (
	"errors"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestStartError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "401 blames the login",
			err:  &deployment.ReplayError{Status: 401, Detail: "Unauthorized"},
			want: []string{"not authorized to replay agent-chat-dev", "logged in"},
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
			got := startError(tt.err, "agent-chat-dev", "0.15.1")
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

func TestCountFinished(t *testing.T) {
	tests := []struct {
		name     string
		statuses []string
		want     int
	}{
		{name: "no batches", statuses: nil, want: 0},
		{name: "all running", statuses: []string{"running", "running"}, want: 0},
		{name: "mixed", statuses: []string{"passed", "running", "failed", "error"}, want: 3},
		{name: "all done", statuses: []string{"passed", "failed"}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &deployment.ReplayRun{}
			for _, s := range tt.statuses {
				run.Batches = append(run.Batches, deployment.ReplayBatch{Status: s})
			}
			if got := countFinished(run); got != tt.want {
				t.Errorf("countFinished() = %d, want %d", got, tt.want)
			}
		})
	}
}
