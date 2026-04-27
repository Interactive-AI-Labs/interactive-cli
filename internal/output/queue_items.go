package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var queueItemColumnMap = map[string]struct {
	Header string
	Value  func(q *clients.QueueItemInfo) string
}{
	"id":          {"ID", func(q *clients.QueueItemInfo) string { return q.ID }},
	"object_id":   {"OBJECT ID", func(q *clients.QueueItemInfo) string { return q.ObjectID }},
	"object_type": {"OBJECT TYPE", func(q *clients.QueueItemInfo) string { return q.ObjectType }},
	"status":      {"STATUS", func(q *clients.QueueItemInfo) string { return q.Status }},
	"completed_at": {
		"COMPLETED AT",
		func(q *clients.QueueItemInfo) string { return LocalTime(q.CompletedAt) },
	},
	"created_at": {
		"CREATED AT",
		func(q *clients.QueueItemInfo) string { return LocalTime(q.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(q *clients.QueueItemInfo) string { return LocalTime(q.UpdatedAt) },
	},
}

func PrintQueueItemList(
	out io.Writer,
	items []clients.QueueItemInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(items) == 0 {
		fmt.Fprintln(out, "No queue items found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := queueItemColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(items))
	for i, item := range items {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := queueItemColumnMap[col]; ok {
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

func PrintQueueItemDetail(out io.Writer, item *clients.QueueItemInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", item.ID)
	fmt.Fprintf(w, "Object ID:\t%s\n", item.ObjectID)
	fmt.Fprintf(w, "Object Type:\t%s\n", item.ObjectType)
	fmt.Fprintf(w, "Status:\t%s\n", item.Status)
	if item.CompletedAt != "" {
		fmt.Fprintf(w, "Completed At:\t%s\n", LocalTime(item.CompletedAt))
	}
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(item.CreatedAt))
	fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(item.UpdatedAt))
	return w.Flush()
}

func PrintQueueItemCreateResult(out io.Writer, item *clients.QueueItemInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Created queue item %q.\n", item.ID)
	fmt.Fprintf(w, "ID:\t%s\n", item.ID)
	fmt.Fprintf(w, "Object ID:\t%s\n", item.ObjectID)
	fmt.Fprintf(w, "Object Type:\t%s\n", item.ObjectType)
	fmt.Fprintf(w, "Status:\t%s\n", item.Status)
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(item.CreatedAt))
	return w.Flush()
}

func PrintQueueItemUpdateResult(out io.Writer, item *clients.QueueItemInfo) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Updated queue item %q.\n", item.ID)
	fmt.Fprintf(w, "Status:\t%s\n", item.Status)
	fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(item.UpdatedAt))
	return w.Flush()
}
