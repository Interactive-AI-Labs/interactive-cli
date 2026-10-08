package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Cobra keeps this annotation key unexported.
const oneRequiredAnnotation = "cobra_annotation_one_required"

// Inputs and version guards, shown wherever they exist.
var indexNamedFlags = map[string]bool{
	"file": true, "env": true, "secret": true,
	"schema-version": true, "message": true, "expect-revision": true,
}

// Every other flag on a list command is a filter.
var listSkipFlags = map[string]bool{
	"organization": true, "project": true, "columns": true, "watch": true, "help": true,
}

func commandIndex(root *cobra.Command, version string) string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"%s v%s: `%s <group> <subcommand> [flags]`. Commands not listed do not exist.\n",
		root.Name(), version, root.Name(),
	)
	b.WriteString(
		"Key flags only; `<command> --help` for the rest. `*` = --json or --yaml (not with --columns); " +
			"commands without `*` or --json print text only.\n",
	)
	for _, group := range indexableChildren(root) {
		if !group.HasAvailableSubCommands() {
			fmt.Fprintf(&b, "%s (%s)%s\n", group.Name(), shortOf(group), bracket(indexFlags(group)))
			continue
		}
		var subs []string
		for _, leaf := range indexLeaves(group) {
			path := strings.TrimPrefix(leaf.CommandPath(), group.CommandPath()+" ")
			subs = append(subs, path+bracket(indexFlags(leaf)))
		}
		fmt.Fprintf(&b, "%s (%s): %s\n", group.Name(), shortOf(group), strings.Join(subs, " · "))
	}
	return b.String()
}

func shortOf(c *cobra.Command) string {
	short, _, _ := strings.Cut(c.Short, " (")
	return short
}

func indexableChildren(c *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, child := range c.Commands() {
		if child.Hidden || child.Deprecated != "" || child.Name() == "help" ||
			child.Name() == "completion" {
			continue
		}
		out = append(out, child)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func indexLeaves(c *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, child := range indexableChildren(c) {
		if child.Runnable() {
			out = append(out, child)
		}
		out = append(out, indexLeaves(child)...)
	}
	return out
}

func bracket(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return " [" + strings.Join(items, " ") + "]"
}

func indexFlags(c *cobra.Command) []string {
	// LocalFlags merges the inherited flags into Flags.
	local, all := c.LocalFlags(), c.Flags()
	var out []string
	shown := map[string]bool{}

	all.VisitAll(func(f *pflag.Flag) {
		if !f.Hidden && len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0 {
			out = append(out, "--"+f.Name+" (required)")
			shown[f.Name] = true
		}
	})

	groups := map[string]bool{}
	all.VisitAll(func(f *pflag.Flag) {
		for _, group := range f.Annotations[oneRequiredAnnotation] {
			names := strings.Fields(group)
			sort.Strings(names)
			text := "(one of --" + strings.Join(names, ", --") + ")"
			if groups[text] {
				continue
			}
			groups[text] = true
			for _, n := range names {
				shown[n] = true
			}
			out = append(out, text)
		}
	})

	local.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || shown[f.Name] || f.Name == "json" || f.Name == "yaml" {
			return
		}
		if indexNamedFlags[f.Name] || (c.Name() == "list" && !listSkipFlags[f.Name]) {
			out = append(out, "--"+f.Name)
		}
	})

	switch {
	case local.Lookup("json") != nil && local.Lookup("yaml") != nil:
		out = append(out, "*")
	case local.Lookup("json") != nil:
		out = append(out, "--json")
	}
	return out
}
