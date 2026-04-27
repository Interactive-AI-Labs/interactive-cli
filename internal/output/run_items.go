package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var runItemColumnMap = map[string]struct {
	Header string
	Value  func(r *clients.DatasetRunItemInfo) string
}{
	"id": {"ID", func(r *clients.DatasetRunItemInfo) string { return r.ID }},
	"dataset_run_name": {
		"RUN NAME",
		func(r *clients.DatasetRunItemInfo) string { return r.DatasetRunName },
	},
	"dataset_item_id": {
		"DATASET ITEM ID",
		func(r *clients.DatasetRunItemInfo) string { return r.DatasetItemID },
	},
	"trace_id": {
		"TRACE ID",
		func(r *clients.DatasetRunItemInfo) string { return r.TraceID },
	},
	"observation_id": {
		"OBSERVATION ID",
		func(r *clients.DatasetRunItemInfo) string { return r.ObservationID },
	},
	"created_at": {
		"CREATED AT",
		func(r *clients.DatasetRunItemInfo) string { return LocalTime(r.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(r *clients.DatasetRunItemInfo) string { return LocalTime(r.UpdatedAt) },
	},
}

func PrintRunItemList(
	out io.Writer,
	items []clients.DatasetRunItemInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(items) == 0 {
		fmt.Fprintln(out, "No run items found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := runItemColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(items))
	for i, item := range items {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := runItemColumnMap[col]; ok {
				row[j] = def.Value(&item)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintRunItemCreateResult(out io.Writer, item *clients.DatasetRunItemInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Created run item %q.\n", item.ID)
	fmt.Fprintf(w, "ID:\t%s\n", item.ID)
	fmt.Fprintf(w, "Run Name:\t%s\n", item.DatasetRunName)
	fmt.Fprintf(w, "Dataset Item ID:\t%s\n", item.DatasetItemID)
	if item.TraceID != "" {
		fmt.Fprintf(w, "Trace ID:\t%s\n", item.TraceID)
	}
	if item.ObservationID != "" {
		fmt.Fprintf(w, "Observation ID:\t%s\n", item.ObservationID)
	}
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(item.CreatedAt))
	return w.Flush()
}
