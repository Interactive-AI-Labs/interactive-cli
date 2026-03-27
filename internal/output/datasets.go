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
		headers[i] = datasetColumnMap[col].Header
	}

	rows := make([][]string, len(datasets))
	for i, ds := range datasets {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = datasetColumnMap[col].Value(&ds)
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
	fmt.Fprintf(out, "ID:          %s\n", ds.ID)
	fmt.Fprintf(out, "Name:        %s\n", ds.Name)
	fmt.Fprintf(out, "Description: %s\n", ds.Description)
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(ds.CreatedAt))
	fmt.Fprintf(out, "Updated At:  %s\n", LocalTime(ds.UpdatedAt))
	if len(ds.Metadata) > 0 && string(ds.Metadata) != "null" {
		fmt.Fprintf(out, "Metadata:    %s\n", string(ds.Metadata))
	}
	return nil
}

func PrintDatasetCreateResult(out io.Writer, ds *clients.DatasetInfo) error {
	fmt.Fprintf(out, "Created dataset %q.\n", ds.Name)
	fmt.Fprintf(out, "ID:          %s\n", ds.ID)
	fmt.Fprintf(out, "Name:        %s\n", ds.Name)
	if ds.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", ds.Description)
	}
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(ds.CreatedAt))
	return nil
}
