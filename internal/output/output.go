package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// NewDescribeWriter returns a tabwriter.Writer for describe/detail output.
// Use \t between label and value — never manual spaces. Call Flush() when done.
func NewDescribeWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.StripEscape)
}

// PrintTable prints data in a tabular format using text/tabwriter.
// Cell text wrapped in '\xff' (tabwriter.Escape) is passed through unchanged
// and ignored when computing column widths — use it to embed ANSI color codes
// without breaking alignment.
func PrintTable(out io.Writer, headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.StripEscape)

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

// LocalTime converts a timestamp to the user's local timezone.
func LocalTime(s string) string {
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000000",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Local().Format(time.RFC1123Z)
		}
	}
	return s
}

// PrintPageMeta prints the standard "Page X of Y (Z total items)" footer.
func PrintPageMeta(out io.Writer, page, totalPages, totalItems int) {
	fmt.Fprintf(out, "\nPage %d of %d (%d total items)\n", page, totalPages, totalItems)
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
