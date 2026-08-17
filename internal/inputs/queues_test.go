package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateQueueListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    clients.AnnotationQueueListOptions
		wantErr bool
	}{
		{"defaults", clients.AnnotationQueueListOptions{Page: 1}, false},
		{"invalid page", clients.AnnotationQueueListOptions{Page: 0}, true},
		{"negative limit", clients.AnnotationQueueListOptions{Page: 1, Limit: -1}, true},
		{
			"valid sort",
			clients.AnnotationQueueListOptions{Page: 1, SortBy: "name", SortOrder: "asc"},
			false,
		},
		{
			"count column sort",
			clients.AnnotationQueueListOptions{Page: 1, SortBy: "count_pending_items"},
			false,
		},
		{
			"unknown sort field",
			clients.AnnotationQueueListOptions{Page: 1, SortBy: "status"},
			true,
		},
		{
			"unknown sort order",
			clients.AnnotationQueueListOptions{Page: 1, SortBy: "name", SortOrder: "ascending"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueueListOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueueListOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
