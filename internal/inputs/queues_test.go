package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestValidateQueueListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    platform.AnnotationQueueListOptions
		wantErr bool
	}{
		{"defaults", platform.AnnotationQueueListOptions{Page: 1}, false},
		{"invalid page", platform.AnnotationQueueListOptions{Page: 0}, true},
		{"negative limit", platform.AnnotationQueueListOptions{Page: 1, Limit: -1}, true},
		{
			"sorting is delegated to the server",
			platform.AnnotationQueueListOptions{Page: 1, SortBy: "bogus", SortOrder: "sideways"},
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
