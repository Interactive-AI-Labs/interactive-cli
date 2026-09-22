package deployment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLocalAgentClientStartReplay(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantRunID  string
		wantErr    string
		wantStatus int
	}{
		{
			name:      "a run id is returned",
			status:    http.StatusAccepted,
			body:      `{"run_id":"run-1"}`,
			wantRunID: "run-1",
		},
		{
			name:    "a start without a run id is refused",
			status:  http.StatusAccepted,
			body:    `{}`,
			wantErr: "no run_id",
		},
		{
			name:    "an unreadable body is refused",
			status:  http.StatusAccepted,
			body:    `not json`,
			wantErr: "failed to decode replay run",
		},
		{
			name:       "the agent's refusal carries its status",
			status:     http.StatusUnauthorized,
			body:       `{"detail":"bad api key"}`,
			wantErr:    "bad api key",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					_, _ = w.Write([]byte(tt.body))
				}),
			)
			defer srv.Close()

			got, err := NewLocalAgentClient(srv.URL+"/", "local-key").
				StartReplay(context.Background(), "", "", "a1", ReplayStartRequest{Repeat: 1})

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("StartReplay: %v", err)
				}
				if got.RunID != tt.wantRunID {
					t.Errorf("run id = %q, want %q", got.RunID, tt.wantRunID)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to mention %q", err, tt.wantErr)
			}
			if tt.wantStatus != 0 {
				var re *ReplayError
				if !errors.As(err, &re) || re.Status != tt.wantStatus {
					t.Errorf("error = %v, want a *ReplayError with status %d", err, tt.wantStatus)
				}
			}
		})
	}
}

// The agent serves no status stream, so a local run polls and never asks for one.
func TestLocalAgentClientFollowReplay(t *testing.T) {
	tests := []struct {
		name         string
		statuses     []string
		wantStatus   string
		wantProgress int
	}{
		{
			name:         "a run already finished is read once",
			statuses:     []string{ReplayStatusPassed},
			wantStatus:   ReplayStatusPassed,
			wantProgress: 1,
		},
		{
			name:         "a running run is polled until it settles",
			statuses:     []string{ReplayStatusRunning, ReplayStatusRunning, ReplayStatusFailed},
			wantStatus:   ReplayStatusFailed,
			wantProgress: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPaths []string
			var gotAuth string
			polls := 0

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotPaths = append(gotPaths, r.URL.RequestURI())
					gotAuth = r.Header.Get("Authorization")
					status := tt.statuses[min(polls, len(tt.statuses)-1)]
					polls++
					_, _ = w.Write([]byte(`{"run_id":"run-1","status":"` + status + `"}`))
				}),
			)
			defer srv.Close()

			c := NewLocalAgentClient(srv.URL, "local-key")
			c.poll = time.Millisecond

			progress := 0
			run, err := c.FollowReplay(context.Background(), "", "", "a1", "run-1",
				func(*ReplayRun) { progress++ })
			if err != nil {
				t.Fatalf("FollowReplay: %v", err)
			}

			if run.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", run.Status, tt.wantStatus)
			}
			if progress != tt.wantProgress {
				t.Errorf("progress reported %d times, want %d", progress, tt.wantProgress)
			}
			if gotAuth != "Bearer local-key" {
				t.Errorf("auth = %q, want the agent's own key", gotAuth)
			}
			for _, p := range gotPaths {
				if p != "/replays/run-1" {
					t.Errorf("requested %q, want /replays/run-1 with no query", p)
				}
			}
		})
	}
}

// A local agent has no release, so describe reports only what is knowable.
func TestLocalAgentClientDescribeAgent(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		wantVersion string
	}{
		{
			name:        "the version names where the agent is",
			baseURL:     "http://127.0.0.1:8080",
			wantVersion: "local agent at http://127.0.0.1:8080",
		},
		{
			name:        "a trailing slash is not carried into it",
			baseURL:     "http://127.0.0.1:8080/",
			wantVersion: "local agent at http://127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLocalAgentClient(tt.baseURL, "k").
				DescribeAgent(context.Background(), "", "", "my-agent")
			if err != nil {
				t.Fatalf("DescribeAgent: %v", err)
			}
			if got.Version != tt.wantVersion {
				t.Errorf("version = %q, want %q", got.Version, tt.wantVersion)
			}
			if got.Revision != 0 {
				t.Errorf("revision = %d, want 0", got.Revision)
			}
		})
	}
}

// The experiment flags are only useful if they reach the agent, so assert the body.
func TestReplayStartRequestBody(t *testing.T) {
	tests := []struct {
		name string
		req  ReplayStartRequest
		want string
	}{
		{
			name: "omitted when unset",
			req:  ReplayStartRequest{Repeat: 1, Concurrency: 8},
			want: `{"repeat":1,"concurrency":8}`,
		},
		{
			name: "carried when named",
			req: ReplayStartRequest{
				Repeat: 1, Concurrency: 8, ExperimentName: "prompt-v4 sweep",
			},
			want: `{"repeat":1,"concurrency":8,"experiment_name":"prompt-v4 sweep"}`,
		},
		{
			name: "reuse rides along",
			req: ReplayStartRequest{
				Repeat: 1, Concurrency: 8,
				ExperimentName: "prompt-v4 sweep", ExperimentNameReuse: true,
			},
			want: `{"repeat":1,"concurrency":8,"experiment_name":"prompt-v4 sweep","experiment_name_reuse":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("body  = %s\nwant  = %s", got, tt.want)
			}
		})
	}
}
