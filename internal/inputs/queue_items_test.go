package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

func TestValidateQueueItemListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    api.QueueItemListOptions
		wantErr bool
	}{
		{"defaults", api.QueueItemListOptions{Page: 1}, false},
		{"invalid page", api.QueueItemListOptions{Page: 0}, true},
		{"negative limit", api.QueueItemListOptions{Page: 1, Limit: -1}, true},
		{
			"sorting is delegated to the server",
			api.QueueItemListOptions{Page: 1, SortBy: "bogus", SortOrder: "sideways"},
			false,
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
