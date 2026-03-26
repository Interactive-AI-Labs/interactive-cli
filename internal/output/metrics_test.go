package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintMetricsDaily(t *testing.T) {
	traceCount := 4
	observationCount := 12
	totalCost := 0.05
	totalTokens := 300
	modelObs := 10

	tests := []struct {
		name       string
		metrics    []clients.DailyMetric
		meta       clients.PageMeta
		columns    []string
		showModels bool
		want       string
	}{
		{
			name:    "empty list",
			columns: inputs.DefaultMetricsDailyColumns,
			want:    "No daily metrics found.\n",
		},
		{
			name: "list with models",
			metrics: []clients.DailyMetric{{
				Date:              "2025-01-01",
				CountTraces:       &traceCount,
				CountObservations: &observationCount,
				TotalCost:         &totalCost,
				TotalTokens:       &totalTokens,
				Models: []clients.ModelUsage{{
					Model:             "gpt-4",
					CountObservations: &modelObs,
					TotalTokens:       &totalTokens,
					TotalCost:         &totalCost,
				}},
			}},
			meta:       clients.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns:    inputs.DefaultMetricsDailyColumns,
			showModels: true,
			want: "DATE         TRACE COUNT   OBSERVATION COUNT   TOTAL COST\n" +
				"2025-01-01   4             12                  $0.050000\n" +
				"\nModel Breakdown for 2025-01-01\n" +
				"MODEL   OBSERVATIONS   TOTAL TOKENS   TOTAL COST\n" +
				"gpt-4   10             300            $0.050000\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintMetricsDaily(&buf, tt.metrics, tt.meta, tt.columns, tt.showModels)
			if err != nil {
				t.Fatalf("PrintMetricsDaily() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
