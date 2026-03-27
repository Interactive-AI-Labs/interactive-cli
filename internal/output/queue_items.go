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
	fmt.Fprintf(out, "ID:           %s\n", item.ID)
	fmt.Fprintf(out, "Object ID:    %s\n", item.ObjectID)
	fmt.Fprintf(out, "Object Type:  %s\n", item.ObjectType)
	fmt.Fprintf(out, "Status:       %s\n", item.Status)
	if item.CompletedAt != "" {
		fmt.Fprintf(out, "Completed At: %s\n", LocalTime(item.CompletedAt))
	}
	fmt.Fprintf(out, "Created At:   %s\n", LocalTime(item.CreatedAt))
	fmt.Fprintf(out, "Updated At:   %s\n", LocalTime(item.UpdatedAt))
	return nil
}

func PrintQueueItemCreateResult(out io.Writer, item *clients.QueueItemInfo) error {
	fmt.Fprintf(out, "Created queue item %q.\n", item.ID)
	fmt.Fprintf(out, "ID:          %s\n", item.ID)
	fmt.Fprintf(out, "Object ID:   %s\n", item.ObjectID)
	fmt.Fprintf(out, "Object Type: %s\n", item.ObjectType)
	fmt.Fprintf(out, "Status:      %s\n", item.Status)
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(item.CreatedAt))
	return nil
}

func PrintQueueItemUpdateResult(out io.Writer, item *clients.QueueItemInfo) error {
	fmt.Fprintf(out, "Updated queue item %q.\n", item.ID)
	fmt.Fprintf(out, "Status:      %s\n", item.Status)
	fmt.Fprintf(out, "Updated At:  %s\n", LocalTime(item.UpdatedAt))
	return nil
}
