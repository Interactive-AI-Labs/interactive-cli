package inputs

import (
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
		{"all defaults", clients.TraceListOptions{Page: 1}, false},
		{"valid page and limit", clients.TraceListOptions{Page: 1, Limit: 50}, false},
		{"zero page", clients.TraceListOptions{Page: 0}, true},
		{"negative page", clients.TraceListOptions{Page: -1}, true},
		{"negative limit", clients.TraceListOptions{Page: 1, Limit: -1}, true},
		{
			"valid from-timestamp",
			clients.TraceListOptions{Page: 1, FromTimestamp: "2025-01-01T00:00:00Z"},
			false,
		},
		{
			"valid to-timestamp",
			clients.TraceListOptions{Page: 1, ToTimestamp: "2025-12-31T23:59:59Z"},
			false,
		},
		{
			"valid timestamp with offset",
			clients.TraceListOptions{Page: 1, FromTimestamp: "2025-01-01T00:00:00+02:00"},
			false,
		},
		{
			"invalid from-timestamp",
			clients.TraceListOptions{Page: 1, FromTimestamp: "not-a-date"},
			true,
		},
		{
			"invalid to-timestamp",
			clients.TraceListOptions{Page: 1, ToTimestamp: "2025-01-01"},
			true,
		},
		// Enum values (order-by, order, level, fields) are passed through to
		// the server for validation — no client-side checks.
		{
			"order-by passed through",
			clients.TraceListOptions{Page: 1, OrderBy: "anything"},
			false,
		},
		{"order passed through", clients.TraceListOptions{Page: 1, Order: "anything"}, false},
		{"level passed through", clients.TraceListOptions{Page: 1, Level: "UNKNOWN"}, false},
		{"fields passed through", clients.TraceListOptions{Page: 1, Fields: "unknown"}, false},
		// Cost filters
		{"negative min-cost", clients.TraceListOptions{Page: 1, MinCost: ptrFloat(-1)}, true},
		{"negative max-cost", clients.TraceListOptions{Page: 1, MaxCost: ptrFloat(-1)}, true},
		{
			"min-cost > max-cost",
			clients.TraceListOptions{Page: 1, MinCost: ptrFloat(5), MaxCost: ptrFloat(1)},
			true,
		},
		{
			"valid cost range",
			clients.TraceListOptions{Page: 1, MinCost: ptrFloat(0.01), MaxCost: ptrFloat(1)},
			false,
		},
		// Latency filters
		{"negative min-latency", clients.TraceListOptions{Page: 1, MinLatency: ptrFloat(-1)}, true},
		{"negative max-latency", clients.TraceListOptions{Page: 1, MaxLatency: ptrFloat(-1)}, true},
		{
			"min-latency > max-latency",
			clients.TraceListOptions{Page: 1, MinLatency: ptrFloat(10), MaxLatency: ptrFloat(1)},
			true,
		},
		{
			"valid latency range",
			clients.TraceListOptions{Page: 1, MinLatency: ptrFloat(0), MaxLatency: ptrFloat(60)},
			false,
		},
		// Token filters
		{"negative min-tokens", clients.TraceListOptions{Page: 1, MinTokens: ptrInt(-1)}, true},
		{"negative max-tokens", clients.TraceListOptions{Page: 1, MaxTokens: ptrInt(-1)}, true},
		{
			"min-tokens > max-tokens",
			clients.TraceListOptions{Page: 1, MinTokens: ptrInt(100), MaxTokens: ptrInt(10)},
			true,
		},
		{
			"valid tokens range",
			clients.TraceListOptions{Page: 1, MinTokens: ptrInt(0), MaxTokens: ptrInt(1000)},
			false,
		},
		// Search length
		{"search within limit", clients.TraceListOptions{Page: 1, Search: "hello"}, false},
		{
			"search exceeds limit",
			clients.TraceListOptions{Page: 1, Search: strings.Repeat("a", 201)},
			true,
		},
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

func ptrFloat(v float64) *float64 { return &v }
func ptrInt(v int) *int           { return &v }
