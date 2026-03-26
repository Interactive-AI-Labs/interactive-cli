package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateMetricsDailyColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"default columns", DefaultMetricsDailyColumns, false},
		{"all columns", AllMetricsDailyColumns, false},
		{"unknown", []string{"date", "unknown"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.columns, AllMetricsDailyColumns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetricsDailyColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMetricsDailyOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    clients.MetricsDailyOptions
		wantErr bool
	}{
		{
			"valid",
			clients.MetricsDailyOptions{FromTimestamp: "2025-01-01T00:00:00Z", Page: 1},
			false,
		},
		{
			"invalid timestamp",
			clients.MetricsDailyOptions{FromTimestamp: "2025-01-01", Page: 1},
			true,
		},
		{"invalid page", clients.MetricsDailyOptions{Page: 0}, true},
		{"negative limit", clients.MetricsDailyOptions{Page: 1, Limit: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricsDailyOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetricsDailyOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
