package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintSessionList(t *testing.T) {
	traceCount := 2
	duration := 12.5
	cost := 0.03
	totalTokens := 1200

	tests := []struct {
		name     string
		sessions []clients.SessionInfo
		meta     clients.PageMeta
		columns  []string
		want     string
	}{
		{
			name:    "empty list",
			columns: inputs.DefaultSessionColumns,
			want:    "No sessions found.\n",
		},
		{
			name: "default columns",
			sessions: []clients.SessionInfo{{
				ID:              "sess-1",
				CreatedAt:       "2025-01-01",
				Environment:     "prod",
				TraceCount:      &traceCount,
				DurationSeconds: &duration,
				TotalCost:       &cost,
				TotalTokens:     &totalTokens,
			}},
			meta:    clients.PageMeta{Page: 1, TotalPages: 2, TotalItems: 3},
			columns: inputs.DefaultSessionColumns,
			want: "ID       CREATED AT   ENVIRONMENT   TRACE COUNT   DURATION (s)   TOTAL COST   TOTAL TOKENS\n" +
				"sess-1   2025-01-01   prod          2             12.50s         $0.030000    1200\n" +
				"\nPage 1 of 2 (3 total items)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintSessionList(&buf, tt.sessions, tt.meta, tt.columns)
			if err != nil {
				t.Fatalf("PrintSessionList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintSessionDetail(t *testing.T) {
	traceCount := 2
	duration := 12.5
	cost := 0.03
	inputTokens := 100
	outputTokens := 200
	totalTokens := 300
	latency := 1500.0
	observationCount := 4

	session := &clients.SessionDetail{
		SessionInfo: clients.SessionInfo{
			ID:              "sess-1",
			CreatedAt:       "2025-01-01",
			UpdatedAt:       "2025-01-02",
			Environment:     "prod",
			UserID:          "user-1",
			TraceCount:      &traceCount,
			DurationSeconds: &duration,
			TotalCost:       &cost,
			InputTokens:     &inputTokens,
			OutputTokens:    &outputTokens,
			TotalTokens:     &totalTokens,
		},
		Traces: []clients.SessionTraceSummary{{
			ID:               "trace-1",
			Name:             "chat",
			Timestamp:        "2025-01-01",
			LatencyMs:        &latency,
			TotalCost:        &cost,
			ObservationCount: &observationCount,
			TotalTokens:      &totalTokens,
			Level:            "DEFAULT",
		}},
	}

	want := "ID:          sess-1\n" +
		"Created At:  2025-01-01\n" +
		"Updated At:  2025-01-02\n" +
		"Environment: prod\n" +
		"User ID:     user-1\n" +
		"\n--- Metrics ---\n" +
		"Trace Count:       2\n" +
		"Duration Seconds:  12.50s\n" +
		"Total Cost:        $0.030000\n" +
		"Input Tokens:      100\n" +
		"Output Tokens:     200\n" +
		"Total Tokens:      300\n" +
		"\n--- Trace Summaries ---\n" +
		"ID        NAME   TIMESTAMP    LATENCY   COST        OBSERVATIONS   TOTAL TOKENS   LEVEL\n" +
		"trace-1   chat   2025-01-01   1.50s     $0.030000   4              300            DEFAULT\n"

	var buf bytes.Buffer
	if err := PrintSessionDetail(&buf, session); err != nil {
		t.Fatalf("PrintSessionDetail() error = %v", err)
	}
	if got := buf.String(); got != want {
		t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
}
