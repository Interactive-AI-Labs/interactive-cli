package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandIndex(t *testing.T) {
	tests := []struct {
		leaf  string
		setup func(c *cobra.Command)
		want  string
	}{
		{"get", func(c *cobra.Command) {
			c.Parent().PersistentFlags().String("db", "", "")
			_ = c.Parent().MarkPersistentFlagRequired("db")
		}, "get [--db (required)]"},
		{"get", func(c *cobra.Command) {
			c.Flags().String("a", "", "")
			c.Flags().String("b", "", "")
			c.MarkFlagsOneRequired("b", "a")
		}, "get [(one of --a, --b)]"},
		{"get", func(c *cobra.Command) {
			c.Flags().Bool("json", false, "")
			c.Flags().Bool("yaml", false, "")
		}, "get [*]"},
		{"get", func(c *cobra.Command) { c.Flags().Bool("json", false, "") }, "get [--json]"},
		{"get", func(c *cobra.Command) {
			c.Flags().String("name", "", "")
			c.Flags().String("project", "", "")
			c.Flags().String("hide", "", "")
			_ = c.Flags().MarkHidden("hide")
		}, "get [--name]"},
		{"list", func(c *cobra.Command) {
			c.Flags().String("from", "", "From (ISO 8601, default: 7 days ago)")
			c.Flags().Int("limit", 0, "Items per page (max 100)")
		}, "list [--from (default 7 days ago) --limit (max 100)]"},
		{"logs", func(c *cobra.Command) {
			c.Flags().String("since", "", "Look back (max 72h, e.g. 1h)")
			c.Flags().Bool("follow", false, "Stream")
		}, "logs [--follow --since (max 72h)]"},
		{"get <id> [rev]", func(*cobra.Command) {}, "get <id> [rev]"},
		{"get <id>", func(c *cobra.Command) { indexCommand(c, "(no --x)") }, "get <id> (no --x)"},
		{"get", func(c *cobra.Command) {
			c.Flags().Bool("force", false, "")
			c.Flags().String("env", "", "")
			indexFlag(c, "env", "replaces")
		}, "get [--env (replaces) --force]"},
	}
	for _, tt := range tests {
		root := &cobra.Command{Use: "iai"}
		group := &cobra.Command{Use: "things", Short: "Manage things (tables)"}
		leaf := &cobra.Command{Use: tt.leaf, Run: func(*cobra.Command, []string) {}}
		root.AddCommand(group)
		group.AddCommand(leaf)
		tt.setup(leaf)

		lines := strings.Split(commandIndex(root, "1.2.3"), "\n")
		if got, want := lines[2], "things (Manage things): "+tt.want; got != want {
			t.Errorf("got  %q\nwant %q", got, want)
		}
	}
}

func TestIndexFlagsExist(t *testing.T) {
	if err := errors.Join(indexFlagErrs...); err != nil {
		t.Fatal(err)
	}
}

func TestIndexLiftsHelpTextNotes(t *testing.T) {
	tests := []struct{ command, flag, note string }{
		{"agents logs", "limit", "max 5000"},
		{"agents logs", "since", "max 72h"},
		{"comments list", "limit", "max 100"},
		{"databases logs", "limit", "max 5000"},
		{"databases logs", "since", "max 72h"},
		{"dataset-items list", "limit", "max 100"},
		{"dataset-runs list", "limit", "max 100"},
		{"datasets list", "limit", "max 100"},
		{"jobs logs", "limit", "max 5000"},
		{"jobs logs", "since", "max 72h"},
		{"jobs runs logs", "limit", "max 5000"},
		{"jobs runs logs", "since", "max 72h"},
		{"mcps logs", "limit", "max 5000"},
		{"mcps logs", "since", "max 72h"},
		{"metrics list", "from-timestamp", "default 7 days ago"},
		{"metrics list", "limit", "max 365"},
		{"observations list", "from-timestamp", "default 7 days ago"},
		{"observations list", "limit", "max 100"},
		{"queue-items list", "limit", "max 100"},
		{"queues list", "limit", "max 100"},
		{"replicas logs", "limit", "max 5000"},
		{"replicas logs", "since", "max 72h"},
		{"router models list", "limit", "max 100"},
		{"run-items list", "limit", "max 100"},
		{"score-configs list", "limit", "max 100"},
		{"scores list", "from-timestamp", "default 7 days ago"},
		{"scores list", "limit", "max 100"},
		{"services logs", "limit", "max 5000"},
		{"services logs", "since", "max 72h"},
		{"sessions list", "from-timestamp", "default 7 days ago"},
		{"sessions list", "limit", "max 100"},
		{"traces list", "from-timestamp", "default 7 days ago"},
		{"traces list", "limit", "max 100"},
		{"traces list", "search", "max 200 characters"},
	}
	for _, tt := range tests {
		c, _, err := rootCmd.Find(strings.Fields(tt.command))
		if err != nil {
			t.Fatal(err)
		}
		if got := flagNote(c.Flags().Lookup(tt.flag)); got != tt.note {
			t.Errorf("%s --%s: got %q, want %q", tt.command, tt.flag, got, tt.note)
		}
	}
}
