package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
