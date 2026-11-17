package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestPrepareRouterModelListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    platform.RouterModelListOptions
		want    platform.RouterModelListOptions
		wantErr string
	}{
		{
			name: "defaults",
			opts: platform.RouterModelListOptions{Limit: 50},
			want: platform.RouterModelListOptions{Limit: 50},
		},
		{
			name: "page, limit and search pass through",
			opts: platform.RouterModelListOptions{Page: 2, Limit: 100, Search: "claude"},
			want: platform.RouterModelListOptions{Page: 2, Limit: 100, Search: "claude"},
		},
		{
			name: "global is sent as us",
			opts: platform.RouterModelListOptions{Limit: 50, Region: "global"},
			want: platform.RouterModelListOptions{Limit: 50, Region: "us"},
		},
		{
			name: "us is still accepted",
			opts: platform.RouterModelListOptions{Limit: 50, Region: "us"},
			want: platform.RouterModelListOptions{Limit: 50, Region: "us"},
		},
		{
			name: "eu",
			opts: platform.RouterModelListOptions{Limit: 50, Region: "eu"},
			want: platform.RouterModelListOptions{Limit: 50, Region: "eu"},
		},
		{
			name:    "negative page",
			opts:    platform.RouterModelListOptions{Page: -1, Limit: 50},
			wantErr: "page must be >= 0, got -1",
		},
		{
			name:    "limit zero",
			opts:    platform.RouterModelListOptions{},
			wantErr: "limit must be between 1 and 100, got 0",
		},
		{
			name:    "limit too high",
			opts:    platform.RouterModelListOptions{Limit: 101},
			wantErr: "limit must be between 1 and 100, got 101",
		},
		{
			name:    "unknown region",
			opts:    platform.RouterModelListOptions{Limit: 50, Region: "moon"},
			wantErr: `invalid region "moon": must be one of global, eu`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrepareRouterModelListOptions(tt.opts)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("PrepareRouterModelListOptions() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("PrepareRouterModelListOptions() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("PrepareRouterModelListOptions() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
