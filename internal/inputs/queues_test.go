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
			"sorting is delegated to the server",
			clients.AnnotationQueueListOptions{Page: 1, SortBy: "bogus", SortOrder: "sideways"},
			false,
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
