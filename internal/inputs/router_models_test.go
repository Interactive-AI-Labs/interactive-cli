package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestValidateRouterModelListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    platform.RouterModelListOptions
		wantErr bool
	}{
		{"valid defaults", platform.RouterModelListOptions{Page: 0, Limit: 50}, false},
		{"valid region", platform.RouterModelListOptions{Limit: 100, Region: "eu"}, false},
		{"negative page", platform.RouterModelListOptions{Page: -1, Limit: 50}, true},
		{"limit zero", platform.RouterModelListOptions{Limit: 0}, true},
		{"limit too high", platform.RouterModelListOptions{Limit: 101}, true},
		{"invalid region", platform.RouterModelListOptions{Limit: 50, Region: "apac"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRouterModelListOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRouterModelListOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
