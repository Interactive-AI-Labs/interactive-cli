package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateQueueItemListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    clients.QueueItemListOptions
		wantErr bool
	}{
		{"defaults", clients.QueueItemListOptions{Page: 1}, false},
		{"invalid page", clients.QueueItemListOptions{Page: 0}, true},
		{"negative limit", clients.QueueItemListOptions{Page: 1, Limit: -1}, true},
		{
			"valid sort",
			clients.QueueItemListOptions{Page: 1, SortBy: "completed_at", SortOrder: "desc"},
			false,
		},
		{
			"unknown sort field",
			clients.QueueItemListOptions{Page: 1, SortBy: "name"},
			true,
		},
		{
			"unknown sort order",
			clients.QueueItemListOptions{Page: 1, SortOrder: "down"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueueItemListOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueueItemListOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
