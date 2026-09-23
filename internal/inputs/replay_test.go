package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/google/go-cmp/cmp"
)

func TestLoadScenarioFile(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		content        string
		want           map[string]any
		wantErr        bool
		errContains    string
		useNonexistent bool
	}{
		{
			name:     "yaml defaults the scenario name to the file name",
			filename: "account-lock.yaml",
			content: `customer_id: replay-1
context_variables:
  open_tickets: []
  meta: {}
messages:
  - hi
`,
			want: map[string]any{
				"scenario":    "account-lock",
				"customer_id": "replay-1",
				"context_variables": map[string]any{
					"open_tickets": []any{},
					"meta":         map[string]any{},
				},
				"messages": []any{"hi"},
			},
		},
		{
			name:     "json keeps an explicit scenario name",
			filename: "x.json",
			content:  `{"scenario":"named","customer_id":"replay-2","messages":["a"]}`,
			want: map[string]any{
				"scenario":    "named",
				"customer_id": "replay-2",
				"messages":    []any{"a"},
			},
		},
		{
			name:        "a document that is not an object",
			filename:    "list.yaml",
			content:     "- a\n- b\n",
			wantErr:     true,
			errContains: "must be a YAML or JSON object",
		},
		{
			name:        "unparseable document",
			filename:    "bad.yaml",
			content:     "messages: [unclosed\n",
			wantErr:     true,
			errContains: "failed to parse scenario file",
		},
		{
			name:           "missing file",
			useNonexistent: true,
			wantErr:        true,
			errContains:    "failed to read scenario file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/nonexistent/scenario.yaml"
			if !tt.useNonexistent {
				path = filepath.Join(t.TempDir(), tt.filename)
				if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			got, err := LoadScenarioFile(path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("LoadScenarioFile() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadScenarioFile() unexpected error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LoadScenarioFile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Every option the caller can set has to reach the request, or the flag does nothing.
func TestBuildReplayStartRequest(t *testing.T) {
	body := map[string]any{"customer_id": "replay-1"}

	for _, tt := range []struct {
		name string
		in   ReplayInput
		body map[string]any
		want deployment.ReplayStartRequest
	}{
		{
			name: "a dataset run carries nothing extra",
			in:   ReplayInput{Dataset: "replay-chat", Repeat: 1, Concurrency: 8},
			want: deployment.ReplayStartRequest{
				Dataset: "replay-chat", Repeat: 1, Concurrency: 8,
			},
		},
		{
			name: "an inline scenario travels in the body",
			in:   ReplayInput{Repeat: 1, Concurrency: 8},
			body: body,
			want: deployment.ReplayStartRequest{
				ScenarioBody: body, Repeat: 1, Concurrency: 8,
			},
		},
		{
			name: "a narrowed dataset keeps its scenario names",
			in: ReplayInput{
				Dataset: "replay-chat", Scenarios: []string{"a", "b"}, Repeat: 3, Concurrency: 16,
			},
			want: deployment.ReplayStartRequest{
				Dataset: "replay-chat", Scenarios: []string{"a", "b"}, Repeat: 3, Concurrency: 16,
			},
		},
		{
			name: "the experiment and session options all arrive",
			in: ReplayInput{
				Dataset: "replay-chat", Repeat: 1, Concurrency: 8,
				ExperimentName: "prompt-v4 sweep", ExperimentNameReuse: true, KeepSessions: true,
			},
			want: deployment.ReplayStartRequest{
				Dataset: "replay-chat", Repeat: 1, Concurrency: 8,
				ExperimentName: "prompt-v4 sweep", ExperimentNameReuse: true, KeepSessions: true,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildReplayStartRequest(tt.in, tt.body)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("request mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
