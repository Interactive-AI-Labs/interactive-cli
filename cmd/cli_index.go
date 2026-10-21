package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	commandNoteAnnotation = "index_command_note"
	flagNoteAnnotation    = "index_flag_note"
	jsonShapeAnnotation   = "index_json_shape"
)

// Cobra keeps these annotation keys unexported.
const (
	oneRequiredAnnotation       = "cobra_annotation_one_required"
	mutuallyExclusiveAnnotation = "cobra_annotation_mutually_exclusive"
)

// The header states these flags' conflicts once.
var indexHeaderFlags = map[string]bool{
	"json": true, "yaml": true, "columns": true, "watch": true,
	"raw": true, "decode": true, "fields": true, "all-fields": true,
}

var indexSkipFlags = map[string]bool{
	"organization": true,
	"project":      true,
	"columns":      true,
	"help":         true,
	"json":         true,
	"yaml":         true,
}

var indexFlagErrs []error

// quoted records an error unless help already says note.
func quoted(where, help, note string) bool {
	if strings.Contains(strings.ToLower(help), strings.ToLower(note)) {
		return true
	}
	indexFlagErrs = append(indexFlagErrs, fmt.Errorf("%s: help text lacks %q", where, note))
	return false
}

func indexCommand(c *cobra.Command, note string) {
	if !quoted(c.CommandPath(), c.Long+" "+c.Short, note) {
		return
	}
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[commandNoteAnnotation] = note
}

// indexJSONShape names a group's --json payload, as spelled out in the header.
func indexJSONShape(c *cobra.Command, shape string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[jsonShapeAnnotation] = shape
}

func indexFlag(c *cobra.Command, name, note string) {
	where := c.CommandPath() + " --" + name
	f := c.Flags().Lookup(name)
	if f == nil {
		indexFlagErrs = append(indexFlagErrs, fmt.Errorf("%s: no such flag", where))
		return
	}
	if quoted(where, f.Usage, note) {
		_ = c.Flags().SetAnnotation(name, flagNoteAnnotation, []string{note})
	}
}

func commandIndex(root *cobra.Command, version string) string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"%s v%s: `%s <group> <subcommand> [flags]`. Commands and flags not listed do not exist; "+
			"global flags, --organization, --project and --columns are left out.\n",
		root.Name(), version, root.Name(),
	)
	b.WriteString(
		"`<command> --help` explains each flag. `*` = --json or --yaml (not with --columns); " +
			"only commands marked `*` or --json accept those output flags. " +
			"--follow and --watch stream until stopped; --watch takes neither --json nor --yaml. " +
			"In logs, --raw, --fields and --all-fields exclude each other and --decode takes neither field flag.\n",
	)
	b.WriteString(
		"--json: `env` groups wrap lists as `{success,data:{<items>:[…],meta}}` and single items as " +
			"`{success,data:{<item>:{…}}}`; `prompts` lists are `{prompts:[…],totalCount}`; the rest is bare. " +
			"Trace and observation input/output are JSON strings (jq `fromjson`).\n",
	)
	for _, group := range indexableChildren(root) {
		if !group.HasAvailableSubCommands() {
			fmt.Fprintf(
				&b,
				"%s (%s)%s\n",
				group.Name()+useArgs(group),
				shortOf(group),
				bracket(indexFlags(group)),
			)
			continue
		}
		var subs []string
		for _, leaf := range indexLeaves(group) {
			path := strings.TrimPrefix(leaf.CommandPath(), group.CommandPath()+" ") + useArgs(leaf)
			if note := leaf.Annotations[commandNoteAnnotation]; note != "" {
				path += " (" + note + ")"
			}
			subs = append(subs, path+bracket(indexFlags(leaf)))
		}
		fmt.Fprintf(&b, "%s (%s): %s\n", group.Name(), shortOf(group), strings.Join(subs, " · "))
	}
	return b.String()
}

func useArgs(c *cobra.Command) string {
	_, args, _ := strings.Cut(c.Use, " ")
	if args == "" {
		return ""
	}
	return " " + args
}

func shortOf(c *cobra.Command) string {
	short, _, _ := strings.Cut(c.Short, " (")
	if shape := c.Annotations[jsonShapeAnnotation]; shape != "" {
		short += "; --json " + shape
	}
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
			note := "required"
			if n := flagNote(f); n != "" {
				note += "; " + n
			}
			out = append(out, "--"+f.Name+" ("+note+")")
			shown[f.Name] = true
		}
	})

	exclusiveGroups := map[string]bool{}
	all.VisitAll(func(f *pflag.Flag) {
		for _, group := range f.Annotations[mutuallyExclusiveAnnotation] {
			exclusiveGroups[groupKey(strings.Fields(group))] = true
		}
	})

	requiredGroups := map[string]bool{}
	all.VisitAll(func(f *pflag.Flag) {
		for _, group := range f.Annotations[oneRequiredAnnotation] {
			names := strings.Fields(group)
			key := groupKey(names)
			if requiredGroups[key] {
				continue
			}
			requiredGroups[key] = true
			for _, n := range names {
				shown[n] = true
			}
			label := "one of"
			if exclusiveGroups[key] {
				label = "exactly one of"
			}
			out = append(out, "("+label+" --"+key+")")
		}
	})

	var exclusive []string
	var pairs [][2]string
	seen := map[string]bool{}
	all.VisitAll(func(f *pflag.Flag) {
		for _, group := range f.Annotations[mutuallyExclusiveAnnotation] {
			if requiredGroups[groupKey(strings.Fields(group))] {
				continue
			}
			names := conflictNames(all, group)
			if len(names) < 2 || allHeaderFlags(names) {
				continue
			}
			joined := strings.Join(names, ", --")
			if seen[joined] {
				continue
			}
			seen[joined] = true
			if len(names) == 2 {
				pairs = append(pairs, [2]string{names[0], names[1]})
			} else {
				exclusive = append(exclusive, "(only one of --"+joined+")")
			}
		}
	})
	exclusive = append(exclusive, exclusivePairs(pairs)...)

	local.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || shown[f.Name] || indexSkipFlags[f.Name] {
			return
		}
		note := flagNote(f)
		if note != "" {
			note = " (" + note + ")"
		}
		out = append(out, "--"+f.Name+note)
	})

	out = append(out, exclusive...)

	switch {
	case local.Lookup("json") != nil && local.Lookup("yaml") != nil:
		out = append(out, "*")
	case local.Lookup("json") != nil:
		out = append(out, "--json")
	}
	return out
}

// conflictNames returns the group's sorted flags minus hidden and skipped ones; --json/--yaml stay.
func conflictNames(flags *pflag.FlagSet, group string) []string {
	var names []string
	for _, n := range strings.Fields(group) {
		f := flags.Lookup(n)
		if f == nil || f.Hidden || (indexSkipFlags[n] && n != "json" && n != "yaml") {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func groupKey(names []string) string {
	names = append([]string(nil), names...)
	sort.Strings(names)
	return strings.Join(names, ", --")
}

// exclusivePairs folds pairs that share a flag into "(--x excludes --a, --b)".
func exclusivePairs(pairs [][2]string) []string {
	var out []string
	for len(pairs) > 0 {
		count := map[string]int{}
		for _, p := range pairs {
			count[p[0]]++
			count[p[1]]++
		}
		hub := pairs[0][0]
		for name, n := range count {
			if n > count[hub] || (n == count[hub] && name < hub) {
				hub = name
			}
		}
		if count[hub] < 2 {
			for _, p := range pairs {
				out = append(out, "(only one of --"+p[0]+", --"+p[1]+")")
			}
			return out
		}
		var others []string
		var rest [][2]string
		for _, p := range pairs {
			switch hub {
			case p[0]:
				others = append(others, p[1])
			case p[1]:
				others = append(others, p[0])
			default:
				rest = append(rest, p)
			}
		}
		sort.Strings(others)
		out = append(out, "(--"+hub+" excludes --"+strings.Join(others, ", --")+")")
		pairs = rest
	}
	return out
}

func allHeaderFlags(names []string) bool {
	for _, n := range names {
		if !indexHeaderFlags[n] {
			return false
		}
	}
	return true
}

func flagNote(f *pflag.Flag) string {
	if note := f.Annotations[flagNoteAnnotation]; len(note) > 0 {
		return note[0]
	}
	return ""
}
