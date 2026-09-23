package deployment

import (
	"context"
	"strings"
	"testing"
)

// The agent serves no status stream, so the path it is polled on carries no query.
func TestLocalRunPath(t *testing.T) {
	tests := []struct {
		name  string
		runID string
		want  string
	}{
		{name: "a plain id", runID: "run-1", want: "/replays/run-1"},
		{name: "a slash cannot walk out of the route", runID: "a/b", want: "/replays/a%2Fb"},
		{name: "a space is escaped", runID: "a b", want: "/replays/a%20b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := localRunPath(tt.runID)
			if got != tt.want {
				t.Errorf("path = %q, want %q", got, tt.want)
			}
			if strings.Contains(got, "?") {
				t.Errorf("path %q carries a query the agent does not serve", got)
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
