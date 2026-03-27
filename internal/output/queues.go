package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var queueColumnMap = map[string]struct {
	Header string
	Value  func(q *clients.AnnotationQueueInfo) string
}{
	"id":   {"ID", func(q *clients.AnnotationQueueInfo) string { return q.ID }},
	"name": {"NAME", func(q *clients.AnnotationQueueInfo) string { return q.Name }},
	"description": {
		"DESCRIPTION",
		func(q *clients.AnnotationQueueInfo) string { return q.Description },
	},
	"score_config_ids": {
		"SCORE CONFIG IDS",
		func(q *clients.AnnotationQueueInfo) string {
			return strings.Join(q.ScoreConfigIDs, ", ")
		},
	},
	"created_at": {
		"CREATED AT",
		func(q *clients.AnnotationQueueInfo) string { return LocalTime(q.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(q *clients.AnnotationQueueInfo) string { return LocalTime(q.UpdatedAt) },
	},
}

func PrintQueueList(
	out io.Writer,
	queues []clients.AnnotationQueueInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(queues) == 0 {
		fmt.Fprintln(out, "No annotation queues found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = queueColumnMap[col].Header
	}

	rows := make([][]string, len(queues))
	for i, q := range queues {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = queueColumnMap[col].Value(&q)
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintQueueDetail(out io.Writer, q *clients.AnnotationQueueInfo) error {
	fmt.Fprintf(out, "ID:          %s\n", q.ID)
	fmt.Fprintf(out, "Name:        %s\n", q.Name)
	fmt.Fprintf(out, "Description: %s\n", q.Description)
	if len(q.ScoreConfigIDs) > 0 {
		fmt.Fprintf(out, "Score Config IDs: %s\n", strings.Join(q.ScoreConfigIDs, ", "))
	}
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(q.CreatedAt))
	fmt.Fprintf(out, "Updated At:  %s\n", LocalTime(q.UpdatedAt))
	return nil
}

func PrintQueueCreateResult(out io.Writer, q *clients.AnnotationQueueInfo) error {
	fmt.Fprintf(out, "Created annotation queue %q.\n", q.Name)
	fmt.Fprintf(out, "ID:          %s\n", q.ID)
	fmt.Fprintf(out, "Name:        %s\n", q.Name)
	if q.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", q.Description)
	}
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(q.CreatedAt))
	return nil
}
