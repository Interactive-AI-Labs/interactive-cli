package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintPromptList(out io.Writer, prompts []clients.PromptInfo) error {
	if len(prompts) == 0 {
		fmt.Fprintln(out, "No prompts found.")
		return nil
	}

	headers := []string{"NAME", "LABELS", "TAGS", "UPDATED"}
	rows := make([][]string, len(prompts))
	for i, p := range prompts {
		rows[i] = []string{
			p.Name,
			TruncateList(p.Labels, 3),
			TruncateList(p.Tags, 3),
			LocalTime(p.LastUpdatedAt),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintPromptDetail(out io.Writer, prompt *clients.PromptDetail) error {
	fmt.Fprintf(out, "Name:        %s\n", prompt.Name)
	fmt.Fprintf(out, "Version:     %d\n", prompt.Version)

	if len(prompt.Labels) > 0 {
		fmt.Fprintf(out, "Labels:      %s\n", strings.Join(prompt.Labels, ", "))
	}
	if len(prompt.Tags) > 0 {
		fmt.Fprintf(out, "Tags:        %s\n", strings.Join(prompt.Tags, ", "))
	}
	if prompt.CreatedAt != "" {
		fmt.Fprintf(out, "Created At:  %s\n", LocalTime(prompt.CreatedAt))
	}
	if prompt.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At:  %s\n", LocalTime(prompt.UpdatedAt))
	}

	if len(prompt.Prompt) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Content:")
		// Prompt may be a JSON-encoded string — try to unquote it for display
		var s string
		if err := json.Unmarshal(prompt.Prompt, &s); err == nil {
			fmt.Fprint(out, s)
			if len(s) > 0 && s[len(s)-1] != '\n' {
				fmt.Fprintln(out)
			}
		} else {
			fmt.Fprintln(out, string(prompt.Prompt))
		}
	}

	return nil
}
