package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintQueueList(t *testing.T) {
	tests := []struct {
		name    string
		queues  []platform.AnnotationQueueInfo
		meta    platform.PageMeta
		columns []string
		want    string
	}{
		{
			name:    "empty list",
			columns: inputs.DefaultQueueColumns,
			want:    "No annotation queues found.\n",
		},
		{
			name: "item count columns",
			queues: []platform.AnnotationQueueInfo{{
				ID:                  "queue-1",
				Name:                "triage",
				CountCompletedItems: 2,
				CountPendingItems:   0,
			}},
			meta:    platform.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns: []string{"id", "name", "count_completed_items", "count_pending_items"},
			want: "ID        NAME     COMPLETED ITEMS   PENDING ITEMS\n" +
				"queue-1   triage   2                 0\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintQueueList(&buf, tt.queues, tt.meta, tt.columns)
			if err != nil {
				t.Fatalf("PrintQueueList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
