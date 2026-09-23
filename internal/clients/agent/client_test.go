package agent

import (
	"strings"
	"testing"
	"time"
)

// The agent serves no status stream, so the path it is polled on carries no query.
func TestRunPath(t *testing.T) {
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

// A trailing slash would double up against every route the client builds.
func TestNewClientTrimsTheBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{name: "left alone", baseURL: "http://127.0.0.1:8080", want: "http://127.0.0.1:8080"},
		{name: "trailing slash", baseURL: "http://127.0.0.1:8080/", want: "http://127.0.0.1:8080"},
		{name: "several", baseURL: "http://127.0.0.1:8080///", want: "http://127.0.0.1:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.baseURL, "k", time.Second).baseURL; got != tt.want {
				t.Errorf("baseURL = %q, want %q", got, tt.want)
			}
		})
	}
}
