package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var datasetItemColumnMap = map[string]struct {
	Header string
	Value  func(d *clients.DatasetItemInfo) string
}{
	"id":     {"ID", func(d *clients.DatasetItemInfo) string { return d.ID }},
	"status": {"STATUS", func(d *clients.DatasetItemInfo) string { return d.Status }},
	"dataset_name": {
		"DATASET NAME",
		func(d *clients.DatasetItemInfo) string { return d.DatasetName },
	},
	"source_trace_id": {
		"SOURCE TRACE ID",
		func(d *clients.DatasetItemInfo) string { return d.SourceTraceID },
	},
	"source_observation_id": {
		"SOURCE OBSERVATION ID",
		func(d *clients.DatasetItemInfo) string { return d.SourceObservationID },
	},
	"created_at": {
		"CREATED AT",
		func(d *clients.DatasetItemInfo) string { return LocalTime(d.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(d *clients.DatasetItemInfo) string { return LocalTime(d.UpdatedAt) },
	},
}

func PrintDatasetItemList(
	out io.Writer,
	items []clients.DatasetItemInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(items) == 0 {
		fmt.Fprintln(out, "No dataset items found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := datasetItemColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(items))
	for i, item := range items {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := datasetItemColumnMap[col]; ok {
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

func PrintDatasetItemDetail(out io.Writer, item *clients.DatasetItemInfo) error {
	fmt.Fprintf(out, "ID:                    %s\n", item.ID)
	fmt.Fprintf(out, "Status:                %s\n", item.Status)
	fmt.Fprintf(out, "Dataset Name:          %s\n", item.DatasetName)
	fmt.Fprintf(out, "Created At:            %s\n", LocalTime(item.CreatedAt))
	fmt.Fprintf(out, "Updated At:            %s\n", LocalTime(item.UpdatedAt))
	if item.SourceTraceID != "" {
		fmt.Fprintf(out, "Source Trace ID:       %s\n", item.SourceTraceID)
	}
	if item.SourceObservationID != "" {
		fmt.Fprintf(out, "Source Observation ID: %s\n", item.SourceObservationID)
	}
	if len(item.Input) > 0 && string(item.Input) != "null" {
		fmt.Fprintf(out, "Input:                 %s\n", prettyJSON(item.Input, false))
	}
	if len(item.ExpectedOutput) > 0 && string(item.ExpectedOutput) != "null" {
		fmt.Fprintf(out, "Expected Output:       %s\n", prettyJSON(item.ExpectedOutput, false))
	}
	if len(item.Metadata) > 0 && string(item.Metadata) != "null" {
		fmt.Fprintf(out, "Metadata:              %s\n", string(item.Metadata))
	}
	return nil
}

func PrintDatasetItemCreateResult(out io.Writer, item *clients.DatasetItemInfo) error {
	fmt.Fprintf(out, "Created dataset item %q.\n", item.ID)
	fmt.Fprintf(out, "ID:           %s\n", item.ID)
	fmt.Fprintf(out, "Status:       %s\n", item.Status)
	fmt.Fprintf(out, "Dataset Name: %s\n", item.DatasetName)
	fmt.Fprintf(out, "Created At:   %s\n", LocalTime(item.CreatedAt))
	return nil
}
