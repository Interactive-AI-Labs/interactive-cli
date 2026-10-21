package cmd

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
			c.Flags().Int("cpu", 0, "CPU (max 7)")
			_ = c.MarkFlagRequired("cpu")
			indexFlag(c, "cpu", "max 7")
		}, "get [--cpu (required; max 7)]"},
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
			indexFlag(c, "limit", "max 100")
		}, "list [--from --limit (max 100)]"},
		{"logs", func(c *cobra.Command) {
			c.Flags().String("since", "", "Look back (max 72h, e.g. 1h)")
			c.Flags().Bool("follow", false, "Stream")
			indexFlag(c, "since", "max 72h")
		}, "logs [--follow --since (max 72h)]"},
		{"get <id> [rev]", func(*cobra.Command) {}, "get <id> [rev]"},
		{"get <id>", func(c *cobra.Command) {
			c.Long = "Takes the id only; there is no --x flag."
			indexCommand(c, "no --x")
		}, "get <id> (no --x)"},
		{"get", func(c *cobra.Command) {
			c.Flags().Bool("force", false, "")
			c.Flags().String("env", "", "Replaces the env")
			indexFlag(c, "env", "replaces")
		}, "get [--env (replaces) --force]"},
		{"get", func(c *cobra.Command) {
			c.Flags().Bool("script", false, "")
			c.Flags().Bool("json", false, "")
			c.Flags().Bool("yaml", false, "")
			c.MarkFlagsMutuallyExclusive("json", "yaml", "script")
		}, "get [--script (only one of --json, --script, --yaml) *]"},
		{"get", func(c *cobra.Command) {
			c.Flags().Bool("watch", false, "")
			c.Flags().Bool("json", false, "")
			c.Flags().Bool("yaml", false, "")
			c.MarkFlagsMutuallyExclusive("json", "yaml")
			c.MarkFlagsMutuallyExclusive("watch", "json")
		}, "get [--watch *]"},
		{"get", func(c *cobra.Command) {
			c.Flags().String("a", "", "")
			c.Flags().String("b", "", "")
			c.MarkFlagsOneRequired("a", "b")
			c.MarkFlagsMutuallyExclusive("a", "b")
		}, "get [(one of --a, --b)]"},
		{"get", func(c *cobra.Command) {
			for _, name := range []string{"a", "b", "c", "d", "e"} {
				c.Flags().String(name, "", "")
			}
			c.MarkFlagsMutuallyExclusive("c", "a")
			c.MarkFlagsMutuallyExclusive("a", "b")
			c.MarkFlagsMutuallyExclusive("d", "e")
		}, "get [--a --b --c --d --e (--a excludes --b, --c) (only one of --d, --e)]"},
		{"get", func(c *cobra.Command) {
			for _, name := range []string{"a", "old", "summary", "columns"} {
				c.Flags().String(name, "", "")
			}
			_ = c.Flags().MarkDeprecated("old", "use --a")
			c.MarkFlagsMutuallyExclusive("a", "old")
			c.MarkFlagsMutuallyExclusive("summary", "columns")
		}, "get [--a --summary]"},
	}
	for _, tt := range tests {
		root, leaf := indexFixture(tt.leaf)
		tt.setup(leaf)

		lines := strings.Split(commandIndex(root, "1.2.3"), "\n")
		if got, want := lines[3], "things (Manage things): "+tt.want; got != want {
			t.Errorf("got  %q\nwant %q", got, want)
		}
	}
}

func TestIndexJSONShape(t *testing.T) {
	root, leaf := indexFixture("get")
	indexJSONShape(leaf.Parent(), "env")

	lines := strings.Split(commandIndex(root, "1.2.3"), "\n")
	if got, want := lines[3], "things (Manage things; --json env): get"; got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

// The header states these conflicts once, so they must hold on every command that has both flags.
var headerConflicts = [][2]string{
	{"json", "watch"},
	{"watch", "yaml"},
	{
		"fields",
		"raw",
	},
	{"all-fields", "raw"},
	{"all-fields", "fields"},
	{"decode", "fields"},
	{"all-fields", "decode"},
}

// Conflicts the header covers without the walk below: `*` = --json or --yaml (not with --columns).
var headerOutputConflicts = [][2]string{{"json", "yaml"}, {"columns", "json"}, {"columns", "yaml"}}

// Both ways: every stated conflict holds, and every conflict the index drops is stated.
func TestIndexHeaderConflictsHold(t *testing.T) {
	stated := map[[2]string]bool{}
	for _, pair := range append(headerConflicts, headerOutputConflicts...) {
		stated[pair] = true
	}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		exclusive := map[[2]string]bool{}
		c.Flags().VisitAll(func(f *pflag.Flag) {
			for _, group := range f.Annotations[mutuallyExclusiveAnnotation] {
				names := strings.Fields(group)
				sort.Strings(names)
				for i, a := range names {
					for _, b := range names[i+1:] {
						exclusive[[2]string{a, b}] = true
						if allHeaderFlags(names) && !stated[[2]string{a, b}] {
							t.Errorf(
								"%s: the index drops --%s/--%s but the header doesn't state it",
								c.CommandPath(),
								a,
								b,
							)
						}
					}
				}
			}
		})
		for _, pair := range headerConflicts {
			if c.Flags().Lookup(pair[0]) != nil && c.Flags().Lookup(pair[1]) != nil &&
				!exclusive[pair] {
				t.Errorf(
					"%s: header says --%s and --%s conflict, but they don't",
					c.CommandPath(),
					pair[0],
					pair[1],
				)
			}
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}

// A note the help doesn't say must be dropped, not printed.
func TestIndexNotesRejectUnquoted(t *testing.T) {
	tests := []struct {
		name     string
		annotate func(c *cobra.Command)
	}{
		{"unknown flag", func(c *cobra.Command) { indexFlag(c, "nope", "max 100") }},
		{"flag note not in help", func(c *cobra.Command) { indexFlag(c, "limit", "max 100") }},
		{"command note not in help", func(c *cobra.Command) { indexCommand(c, "no --x") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(errs []error) { indexFlagErrs = errs }(indexFlagErrs)
			indexFlagErrs = nil

			root, leaf := indexFixture("get")
			leaf.Flags().Int("limit", 0, "Items per page")
			tt.annotate(leaf)

			if len(indexFlagErrs) != 1 {
				t.Fatalf("recorded %d errors, want 1", len(indexFlagErrs))
			}
			lines := strings.Split(commandIndex(root, "1.2.3"), "\n")
			if got, want := lines[3], "things (Manage things): get [--limit]"; got != want {
				t.Errorf("rejected note reached the index\ngot  %q\nwant %q", got, want)
			}
		})
	}
}

func TestIndexAnnotationsMatchFlags(t *testing.T) {
	if err := errors.Join(indexFlagErrs...); err != nil {
		t.Fatal(err)
	}
}

func indexFixture(use string) (root, leaf *cobra.Command) {
	root = &cobra.Command{Use: "iai"}
	group := &cobra.Command{Use: "things", Short: "Manage things (tables)"}
	leaf = &cobra.Command{Use: use, Run: func(*cobra.Command, []string) {}}
	root.AddCommand(group)
	group.AddCommand(leaf)
	return root, leaf
}

// Groups whose --json output has no envelope; every other group with --json needs indexJSONShape.
var bareJSONGroups = map[string]bool{
	"agents": true, "api-keys": true, "collections": true, "databases": true, "images": true,
	"jobs": true, "organizations": true, "projects": true, "replicas": true, "router": true,
	"secrets": true, "services": true, "stacks": true,
}

// Shapes can't be quoted from help, so each group's shape is a decision this test forces.
func TestIndexJSONShapeCoversEveryGroup(t *testing.T) {
	for _, group := range indexableChildren(rootCmd) {
		shaped := group.Annotations[jsonShapeAnnotation] != ""
		if hasJSONLeaf(group) && shaped == bareJSONGroups[group.Name()] {
			t.Errorf(
				"%s: call indexJSONShape or list it in bareJSONGroups, not both or neither",
				group.Name(),
			)
		}
	}
}

func hasJSONLeaf(c *cobra.Command) bool {
	if c.Runnable() && c.LocalFlags().Lookup("json") != nil {
		return true
	}
	for _, child := range c.Commands() {
		if hasJSONLeaf(child) {
			return true
		}
	}
	return false
}
