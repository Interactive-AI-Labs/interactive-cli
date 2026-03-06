package inputs

import (
	"strings"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateTraceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid 32-char hex", "5778886310644bbba99b55ea6a3d40ba", false},
		{"valid UUID with hyphens", "d1c7fb08-4cea-4afb-8d64-e3571bd3902d", false},
		{"valid UUID without hyphens", "d1c7fb084cea4afb8d64e3571bd3902d", false},
		{"valid custom string ID", "my-trace-123", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"too long", strings.Repeat("a", 257), true},
		{"max length is valid", strings.Repeat("a", 256), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTraceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTraceID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTraceListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    clients.TraceListOptions
		wantErr bool
	}{
		{"all defaults", clients.TraceListOptions{}, false},
		{"valid page and limit", clients.TraceListOptions{Page: 1, Limit: 50}, false},
		{"negative page", clients.TraceListOptions{Page: -1}, true},
		{"negative limit", clients.TraceListOptions{Limit: -1}, true},
		{"valid from-timestamp", clients.TraceListOptions{FromTimestamp: "2025-01-01T00:00:00Z"}, false},
		{"valid to-timestamp", clients.TraceListOptions{ToTimestamp: "2025-12-31T23:59:59Z"}, false},
		{"valid timestamp with offset", clients.TraceListOptions{FromTimestamp: "2025-01-01T00:00:00+02:00"}, false},
		{"invalid from-timestamp", clients.TraceListOptions{FromTimestamp: "not-a-date"}, true},
		{"invalid to-timestamp", clients.TraceListOptions{ToTimestamp: "2025-01-01"}, true},
		{"valid order-by", clients.TraceListOptions{OrderBy: "timestamp.desc"}, false},
		{"valid order-by asc", clients.TraceListOptions{OrderBy: "name.asc"}, false},
		{"order-by missing direction", clients.TraceListOptions{OrderBy: "timestamp"}, true},
		{"order-by invalid field", clients.TraceListOptions{OrderBy: "unknown.desc"}, true},
		{"order-by invalid direction", clients.TraceListOptions{OrderBy: "timestamp.up"}, true},
		{"order-by too many parts", clients.TraceListOptions{OrderBy: "a.b.c"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTraceListOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTraceListOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTraceColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"all valid columns", AllTraceColumns, false},
		{"default columns valid", DefaultTraceColumns, false},
		{"single valid column", []string{"id"}, false},
		{"unknown column", []string{"id", "nonexistent"}, true},
		{"empty list is valid", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTraceColumns(tt.columns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTraceColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
