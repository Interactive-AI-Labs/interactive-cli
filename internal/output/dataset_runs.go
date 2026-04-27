package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var datasetRunColumnMap = map[string]struct {
	Header string
	Value  func(r *clients.DatasetRunInfo) string
}{
	"id":   {"ID", func(r *clients.DatasetRunInfo) string { return r.ID }},
	"name": {"NAME", func(r *clients.DatasetRunInfo) string { return r.Name }},
	"description": {
		"DESCRIPTION",
		func(r *clients.DatasetRunInfo) string { return r.Description },
	},
	"dataset_name": {
		"DATASET NAME",
		func(r *clients.DatasetRunInfo) string { return r.DatasetName },
	},
	"created_at": {
		"CREATED AT",
		func(r *clients.DatasetRunInfo) string { return LocalTime(r.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(r *clients.DatasetRunInfo) string { return LocalTime(r.UpdatedAt) },
	},
}

func PrintDatasetRunList(
	out io.Writer,
	runs []clients.DatasetRunInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(runs) == 0 {
		fmt.Fprintln(out, "No dataset runs found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := datasetRunColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(runs))
	for i, run := range runs {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := datasetRunColumnMap[col]; ok {
				row[j] = def.Value(&run)
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

func PrintDatasetRunDetail(out io.Writer, run *clients.DatasetRunInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", run.ID)
	fmt.Fprintf(w, "Name:\t%s\n", run.Name)
	fmt.Fprintf(w, "Dataset Name:\t%s\n", run.DatasetName)
	if run.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", run.Description)
	}
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(run.CreatedAt))
	fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(run.UpdatedAt))
	if len(run.Metadata) > 0 && string(run.Metadata) != "null" {
		fmt.Fprintf(w, "Metadata:\t%s\n", string(run.Metadata))
	}
	return w.Flush()
}
