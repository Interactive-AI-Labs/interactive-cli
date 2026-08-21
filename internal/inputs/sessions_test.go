package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestValidateSessionColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"default columns", DefaultSessionColumns, false},
		{"all columns", AllSessionColumns, false},
		{"unknown column", []string{"id", "unknown"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.columns, AllSessionColumns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSessionListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    platform.SessionListOptions
		wantErr bool
	}{
		{
			"valid",
			platform.SessionListOptions{FromTimestamp: "2025-01-01T00:00:00Z", Page: 1},
			false,
		},
		{
			"invalid timestamp",
			platform.SessionListOptions{FromTimestamp: "2025-01-01", Page: 1},
			true,
		},
		{"invalid page", platform.SessionListOptions{Page: 0}, true},
		{"negative limit", platform.SessionListOptions{Page: 1, Limit: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionListOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSessionListOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
