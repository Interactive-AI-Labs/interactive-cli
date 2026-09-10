package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateReplayInput(t *testing.T) {
	valid := ReplayInput{Dataset: "d", Repeat: 1, Concurrency: 8}
	tests := []struct {
		name    string
		mutate  func(*ReplayInput)
		wantErr string
	}{
		{name: "dataset", mutate: func(*ReplayInput) {}},
		{name: "file", mutate: func(in *ReplayInput) { in.Dataset = ""; in.File = "x.yaml" }},
		{name: "run-id", mutate: func(in *ReplayInput) { in.Dataset = ""; in.RunID = "abc" }},
		{
			name:   "scenarios with dataset",
			mutate: func(in *ReplayInput) { in.Scenarios = []string{"a"} },
		},
		{
			name:    "scenarios only gets the specific message",
			mutate:  func(in *ReplayInput) { in.Dataset = ""; in.Scenarios = []string{"a"} },
			wantErr: "--scenarios requires --dataset",
		},
		{
			name:    "no mode",
			mutate:  func(in *ReplayInput) { in.Dataset = "" },
			wantErr: "one of --dataset, --file, or --run-id is required",
		},
		{
			name:    "scenarios without dataset",
			mutate:  func(in *ReplayInput) { in.Dataset = ""; in.File = "x"; in.Scenarios = []string{"a"} },
			wantErr: "--scenarios requires --dataset",
		},
		{
			name:    "repeat too low",
			mutate:  func(in *ReplayInput) { in.Repeat = 0 },
			wantErr: "--repeat must be between 1 and 20",
		},
		{
			name:    "repeat too high",
			mutate:  func(in *ReplayInput) { in.Repeat = 21 },
			wantErr: "--repeat must be between 1 and 20",
		},
		{
			name:    "concurrency too high",
			mutate:  func(in *ReplayInput) { in.Concurrency = 33 },
			wantErr: "--concurrency must be between 1 and 32",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid
			tt.mutate(&in)
			err := ValidateReplayInput(in)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadScenarioFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("yaml with containers", func(t *testing.T) {
		path := write(
			"account-lock.yaml",
			"customer_id: replay-1\ncontext_variables:\n  open_tickets: []\n  meta: {}\nmessages:\n  - hi\n",
		)
		got, err := LoadScenarioFile(path)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]any{
			"scenario":    "account-lock",
			"customer_id": "replay-1",
			"context_variables": map[string]any{
				"open_tickets": []any{},
				"meta":         map[string]any{},
			},
			"messages": []any{"hi"},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch:\n%s", diff)
		}
	})

	t.Run("json keeps explicit scenario", func(t *testing.T) {
		path := write("x.json", `{"scenario":"named","customer_id":"replay-2","messages":["a"]}`)
		got, err := LoadScenarioFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if got["scenario"] != "named" {
			t.Errorf("scenario = %v", got["scenario"])
		}
	})

	t.Run("not an object", func(t *testing.T) {
		path := write("list.yaml", "- a\n- b\n")
		_, err := LoadScenarioFile(path)
		if err == nil || !strings.Contains(err.Error(), "must be a YAML or JSON object") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadScenarioFile(filepath.Join(dir, "nope.yaml"))
		if err == nil || !strings.Contains(err.Error(), "failed to read scenario file") {
			t.Errorf("error = %v", err)
		}
	})
}
