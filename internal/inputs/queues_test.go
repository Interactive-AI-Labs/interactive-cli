package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

func TestValidateQueueListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    api.AnnotationQueueListOptions
		wantErr bool
	}{
		{"defaults", api.AnnotationQueueListOptions{Page: 1}, false},
		{"invalid page", api.AnnotationQueueListOptions{Page: 0}, true},
		{"negative limit", api.AnnotationQueueListOptions{Page: 1, Limit: -1}, true},
		{
			"sorting is delegated to the server",
			api.AnnotationQueueListOptions{Page: 1, SortBy: "bogus", SortOrder: "sideways"},
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
