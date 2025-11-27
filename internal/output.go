package internal

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
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
