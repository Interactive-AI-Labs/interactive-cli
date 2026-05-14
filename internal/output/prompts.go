package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"znkr.io/diff/textdiff"
)

const colorBlue = "\033[1;34m"

func PrintPromptList(out io.Writer, prompts []clients.PromptInfo) error {
	if len(prompts) == 0 {
		fmt.Fprintln(out, "No prompts found.")
		return nil
	}

	useColor := isTerminal(out)
	headers := []string{"NAME", "LABELS", "TAGS", "UPDATED"}
	rows := make([][]string, len(prompts))
	for i, p := range prompts {
		name := p.Name
		if p.RowType == "folder" {
			name = colorizeFolder(name+"/", useColor)
		}
		rows[i] = []string{
			name,
			TruncateList(p.Labels, 3),
			TruncateList(p.Tags, 3),
			LocalTime(p.LastUpdatedAt),
		}
	}

	return PrintTable(out, headers, rows)
}

// colorizeFolder wraps name in blue ANSI escape codes when color is enabled.
// The codes are bracketed with '\xff' so tabwriter excludes them from column
// width calculations (see PrintTable).
func colorizeFolder(name string, useColor bool) string {
	if !useColor {
		return name
	}
	return "\xff" + colorBlue + "\xff" + name + "\xff" + colorReset + "\xff"
}

func PrintPromptDetail(out io.Writer, prompt *clients.PromptDetail) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", prompt.Name)
	fmt.Fprintf(w, "Version:\t%d\n", prompt.Version)

	if len(prompt.Labels) > 0 {
		fmt.Fprintf(w, "Labels:\t%s\n", strings.Join(prompt.Labels, ", "))
	}
	if len(prompt.Tags) > 0 {
		fmt.Fprintf(w, "Tags:\t%s\n", strings.Join(prompt.Tags, ", "))
	}
	if prompt.CreatedAt != "" {
		fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(prompt.CreatedAt))
	}
	if prompt.UpdatedAt != "" {
		fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(prompt.UpdatedAt))
	}

	if hasConfigPayload(prompt.Config) {
		var indented bytes.Buffer
		if err := json.Indent(&indented, prompt.Config, "", "  "); err == nil {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Config:")
			fmt.Fprintln(w, indented.String())
		}
	}

	if len(prompt.Prompt) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Content:")
		// Prompt may be a JSON-encoded string — try to unquote it for display
		var s string
		if err := json.Unmarshal(prompt.Prompt, &s); err == nil {
			fmt.Fprint(w, s)
			if len(s) > 0 && s[len(s)-1] != '\n' {
				fmt.Fprintln(w)
			}
		} else {
			fmt.Fprintln(w, string(prompt.Prompt))
		}
	}

	return w.Flush()
}

// hasConfigPayload reports whether the prompt's config field carries any
// fields worth rendering. The backend returns "{}" (or omits the field) when
// no config is set; treat both as empty.
func hasConfigPayload(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	return len(m) > 0
}

func PrintPromptVersions(out io.Writer, versions []int) error {
	if len(versions) == 0 {
		fmt.Fprintln(out, "No versions found.")
		return nil
	}

	sorted := make([]int, len(versions))
	copy(sorted, versions)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))

	headers := []string{"VERSION"}
	rows := make([][]string, len(sorted))
	for i, v := range sorted {
		rows[i] = []string{fmt.Sprintf("%d", v)}
	}

	return PrintTable(out, headers, rows)
}

func PrintPromptDiff(
	out io.Writer,
	versionA string,
	a *clients.PromptDetail,
	versionB string,
	b *clients.PromptDetail,
) error {
	contentA := normalizePromptContent(a.Prompt)
	contentB := normalizePromptContent(b.Prompt)

	if contentA == contentB {
		fmt.Fprintln(out, "No differences found.")
		return nil
	}

	var opts []textdiff.Option
	if isTerminal(out) {
		opts = append(opts, textdiff.TerminalColors())
		fmt.Fprintf(out, "%s--- version %s%s\n", colorRed, versionA, colorReset)
		fmt.Fprintf(out, "%s+++ version %s%s\n", colorGreen, versionB, colorReset)
	} else {
		fmt.Fprintf(out, "--- version %s\n", versionA)
		fmt.Fprintf(out, "+++ version %s\n", versionB)
	}

	diff := textdiff.Unified(contentA, contentB, opts...)
	fmt.Fprint(out, diff)
	return nil
}

func normalizePromptContent(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if len(s) > 0 && s[len(s)-1] != '\n' {
			return s + "\n"
		}
		return s
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err == nil && buf.Len() > 0 {
		s := buf.String()
		if s[len(s)-1] != '\n' {
			return s + "\n"
		}
		return s
	}

	return string(raw)
}
