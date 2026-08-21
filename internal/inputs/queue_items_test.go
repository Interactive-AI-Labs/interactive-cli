package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestValidateQueueItemListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    platform.QueueItemListOptions
		wantErr bool
	}{
		{"defaults", platform.QueueItemListOptions{Page: 1}, false},
		{"invalid page", platform.QueueItemListOptions{Page: 0}, true},
		{"negative limit", platform.QueueItemListOptions{Page: 1, Limit: -1}, true},
		{
			"sorting is delegated to the server",
			platform.QueueItemListOptions{Page: 1, SortBy: "bogus", SortOrder: "sideways"},
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
