package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
