package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateRouterModelListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    clients.RouterModelListOptions
		wantErr bool
	}{
		{"valid defaults", clients.RouterModelListOptions{Page: 0, Limit: 50}, false},
		{"valid region", clients.RouterModelListOptions{Limit: 100, Region: "eu"}, false},
		{"negative page", clients.RouterModelListOptions{Page: -1, Limit: 50}, true},
		{"limit zero", clients.RouterModelListOptions{Limit: 0}, true},
		{"limit too high", clients.RouterModelListOptions{Limit: 101}, true},
		{"invalid region", clients.RouterModelListOptions{Limit: 50, Region: "apac"}, true},
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
