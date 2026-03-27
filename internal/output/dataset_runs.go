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
	fmt.Fprintf(out, "ID:           %s\n", run.ID)
	fmt.Fprintf(out, "Name:         %s\n", run.Name)
	fmt.Fprintf(out, "Dataset Name: %s\n", run.DatasetName)
	if run.Description != "" {
		fmt.Fprintf(out, "Description:  %s\n", run.Description)
	}
	fmt.Fprintf(out, "Created At:   %s\n", LocalTime(run.CreatedAt))
	fmt.Fprintf(out, "Updated At:   %s\n", LocalTime(run.UpdatedAt))
	if len(run.Metadata) > 0 && string(run.Metadata) != "null" {
		fmt.Fprintf(out, "Metadata:     %s\n", string(run.Metadata))
	}
	return nil
}
