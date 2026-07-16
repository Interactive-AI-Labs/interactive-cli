package output

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintRouterModelList(
	out io.Writer,
	models []clients.RouterModel,
	meta clients.PageMeta,
) error {
	if len(models) == 0 {
		fmt.Fprintln(out, "No models found.")
		return nil
	}

	headers := []string{"NAME", "CONTEXT", "REGION", "ID"}
	rows := make([][]string, len(models))
	for i, m := range models {
		rows[i] = []string{
			m.ModelName,
			formatInt(m.ContextLength),
			m.Region,
			m.ID,
		}
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintRouterModelDetail(out io.Writer, m *clients.RouterModel) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", m.ID)
	fmt.Fprintf(w, "Model Name:\t%s\n", m.ModelName)
	if m.MarketingName != "" {
		fmt.Fprintf(w, "Marketing:\t%s\n", m.MarketingName)
	}
	fmt.Fprintf(w, "Match:\t%s\n", m.MatchPattern)
	fmt.Fprintf(w, "Region:\t%s\n", m.Region)
	otherRegion := "No"
	if m.HasOtherRegion {
		otherRegion = "Yes"
	}
	fmt.Fprintf(w, "Other Region:\t%s\n", otherRegion)
	if m.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", m.Description)
	}
	if m.ContextLength != nil {
		fmt.Fprintf(w, "Context:\t%s\n", formatInt(m.ContextLength))
	}
	if len(m.Capabilities) > 0 {
		fmt.Fprintf(w, "Capabilities:\t%s\n", strings.Join(m.Capabilities, ", "))
	}
	if m.TokenizerID != "" {
		fmt.Fprintf(w, "Tokenizer:\t%s\n", m.TokenizerID)
	}
	if m.CreatedAt != "" {
		fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(m.CreatedAt))
	}
	if m.LastUsed != "" {
		fmt.Fprintf(w, "Last Used:\t%s\n", LocalTime(m.LastUsed))
	}
	if len(m.Prices) > 0 {
		fmt.Fprintln(w, "Prices:")
		for _, k := range slices.Sorted(maps.Keys(m.Prices)) {
			fmt.Fprintf(w, "  %s:\t%s\n", k, strconv.FormatFloat(m.Prices[k], 'f', -1, 64))
		}
	}
	return w.Flush()
}
