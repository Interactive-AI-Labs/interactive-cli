package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

func TestValidateRouterModelListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    api.RouterModelListOptions
		wantErr bool
	}{
		{"valid defaults", api.RouterModelListOptions{Page: 0, Limit: 50}, false},
		{"valid region", api.RouterModelListOptions{Limit: 100, Region: "eu"}, false},
		{"negative page", api.RouterModelListOptions{Page: -1, Limit: 50}, true},
		{"limit zero", api.RouterModelListOptions{Limit: 0}, true},
		{"limit too high", api.RouterModelListOptions{Limit: 101}, true},
		{"invalid region", api.RouterModelListOptions{Limit: 50, Region: "apac"}, true},
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
