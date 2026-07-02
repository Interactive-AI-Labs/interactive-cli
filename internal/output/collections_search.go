package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// PrintSearchResults renders ranked search hits as a table.
func PrintSearchResults(out io.Writer, r *clients.SearchResponse) error {
	if len(r.Results) == 0 {
		fmt.Fprintln(out, "No results.")
		return nil
	}
	return searchHitsTable(out, r.Results)
}

// PrintBatchSearchResults renders one result block per sub-search.
func PrintBatchSearchResults(out io.Writer, r *clients.BatchSearchResponse) error {
	if len(r.Responses) == 0 {
		fmt.Fprintln(out, "No results.")
		return nil
	}
	for i, resp := range r.Responses {
		fmt.Fprintf(out, "Query %d:\n", i+1)
		if len(resp.Results) == 0 {
			fmt.Fprintln(out, "  No results.")
		} else if err := searchHitsTable(out, resp.Results); err != nil {
			return err
		}
		if i < len(r.Responses)-1 {
			fmt.Fprintln(out)
		}
	}
	return nil
}

func searchHitsTable(out io.Writer, hits []clients.SearchHit) error {
	headers := []string{"SCORE", "ID", "TEXT"}
	rows := make([][]string, len(hits))
	for i, h := range hits {
		rows[i] = []string{fmt.Sprintf("%.4f", h.Score), h.ID, truncateString(h.Text, 60)}
	}
	return PrintTable(out, headers, rows)
}
