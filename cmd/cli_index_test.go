package cmd

import (
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
			c.Flags().String("file", "", "")
			c.Flags().String("name", "", "")
		}, "get [--file]"},
		{"list", func(c *cobra.Command) {
			c.Flags().String("name", "", "")
			c.Flags().String("project", "", "")
		}, "list [--name]"},
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
