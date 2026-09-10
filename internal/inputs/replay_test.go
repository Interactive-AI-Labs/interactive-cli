package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
