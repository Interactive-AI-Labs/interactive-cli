package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var datasetColumnMap = map[string]struct {
	Header string
	Value  func(d *clients.DatasetInfo) string
}{
	"id":          {"ID", func(d *clients.DatasetInfo) string { return d.ID }},
	"name":        {"NAME", func(d *clients.DatasetInfo) string { return d.Name }},
	"description": {"DESCRIPTION", func(d *clients.DatasetInfo) string { return d.Description }},
	"created_at": {
		"CREATED AT",
		func(d *clients.DatasetInfo) string { return LocalTime(d.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(d *clients.DatasetInfo) string { return LocalTime(d.UpdatedAt) },
	},
}

func PrintDatasetList(
	out io.Writer,
	datasets []clients.DatasetInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(datasets) == 0 {
		fmt.Fprintln(out, "No datasets found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := datasetColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(datasets))
	for i, ds := range datasets {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := datasetColumnMap[col]; ok {
				row[j] = def.Value(&ds)
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

func PrintDatasetDetail(out io.Writer, ds *clients.DatasetInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", ds.ID)
	fmt.Fprintf(w, "Name:\t%s\n", ds.Name)
	fmt.Fprintf(w, "Description:\t%s\n", ds.Description)
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(ds.CreatedAt))
	fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(ds.UpdatedAt))
	if len(ds.Metadata) > 0 && string(ds.Metadata) != "null" {
		fmt.Fprintf(w, "Metadata:\t%s\n", string(ds.Metadata))
	}
	return w.Flush()
}

func PrintDatasetCreateResult(out io.Writer, ds *clients.DatasetInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Created dataset %q.\n", ds.Name)
	fmt.Fprintf(w, "ID:\t%s\n", ds.ID)
	fmt.Fprintf(w, "Name:\t%s\n", ds.Name)
	if ds.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", ds.Description)
	}
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(ds.CreatedAt))
	return w.Flush()
}
