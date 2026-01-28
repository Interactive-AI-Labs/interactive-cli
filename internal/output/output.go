package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// PrintTable prints data in a tabular format using text/tabwriter.
func PrintTable(out io.Writer, headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)

	if len(headers) > 0 {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
	}

	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

// TruncateList formats a string slice, showing up to maxVisible items
// with a "(+N more)" suffix if truncated.
func TruncateList(items []string, maxVisible int) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) <= maxVisible {
		return strings.Join(items, ", ")
	}
	visible := strings.Join(items[:maxVisible], ", ")
	return fmt.Sprintf("%s (+%d more)", visible, len(items)-maxVisible)
}

func PrintLoadingDots(out io.Writer) chan struct{} {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Fprint(out, ".")
			}
		}
	}()

	return done
}
