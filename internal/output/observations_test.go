package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintObservationList(t *testing.T) {
	cost := 0.0081
	latency := 9321.0
	tokens := 7474

	tests := []struct {
		name         string
		observations []clients.ObservationInfo
		columns      []string
		want         string
	}{
		{
			name:         "empty list prints message",
			observations: []clients.ObservationInfo{},
			columns:      inputs.DefaultObservationColumns,
			want:         "No observations found.\n",
		},
		{
			name:         "nil list prints message",
			observations: nil,
			columns:      inputs.DefaultObservationColumns,
			want:         "No observations found.\n",
		},
		{
			name: "default columns",
			observations: []clients.ObservationInfo{
				{
					ID:          "obs-123",
					Type:        "GENERATION",
					Name:        "ChatGPT",
					Model:       "gpt-4",
					LatencyMs:   &latency,
					TotalCost:   &cost,
					TotalTokens: &tokens,
				},
			},
			columns: inputs.DefaultObservationColumns,
			want: "ID        TYPE         NAME      MODEL   LATENCY (ms)   COST        TOTAL TOKENS\n" +
				"obs-123   GENERATION   ChatGPT   gpt-4   9321.00ms      $0.008100   7474\n",
		},
		{
			name: "custom columns subset",
			observations: []clients.ObservationInfo{
				{
					ID:      "obs-123",
					TraceID: "trace-456",
					Type:    "SPAN",
				},
			},
			columns: []string{"id", "trace_id", "type"},
			want: "ID        TRACE ID    TYPE\n" +
				"obs-123   trace-456   SPAN\n",
		},
		{
			name: "nil metrics show dash",
			observations: []clients.ObservationInfo{
				{
					ID:   "obs-123",
					Name: "test",
				},
			},
			columns: []string{"id", "latency_ms", "total_cost", "total_tokens"},
			want: "ID        LATENCY (ms)   COST   TOTAL TOKENS\n" +
				"obs-123   -              -      -\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintObservationList(&buf, tt.observations, tt.columns)
			if err != nil {
				t.Fatalf("PrintObservationList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintObservationDetail(t *testing.T) {
	latency := 9321.0
	inputTokens := 6472
	outputTokens := 0
	totalTokens := 7474
	cost := 0.00809
	promptVersion := 3

	tests := []struct {
		name string
		obs  *clients.ObservationDetail
		want string
	}{
		{
			name: "core fields and metrics",
			obs: &clients.ObservationDetail{
				ObservationInfo: clients.ObservationInfo{
					ID:           "obs-123",
					TraceID:      "trace-456",
					Type:         "GENERATION",
					Name:         "ChatGPT",
					StartTime:    "2025-01-01",
					EndTime:      "2025-01-01",
					Level:        "DEFAULT",
					Model:        "gpt-4",
					LatencyMs:    &latency,
					InputTokens:  &inputTokens,
					OutputTokens: &outputTokens,
					TotalTokens:  &totalTokens,
					TotalCost:    &cost,
				},
			},
			want: "ID:           obs-123\n" +
				"Trace ID:     trace-456\n" +
				"Type:         GENERATION\n" +
				"Name:         ChatGPT\n" +
				"Start Time:   2025-01-01\n" +
				"End Time:     2025-01-01\n" +
				"Level:        DEFAULT\n" +
				"\n--- Model ---\n" +
				"Model:   gpt-4\n" +
				"\n--- Metrics ---\n" +
				"Latency:         9321.00ms\n" +
				"Input Tokens:    6472\n" +
				"Output Tokens:   0\n" +
				"Total Tokens:    7474\n" +
				"Total Cost:      $0.008090\n",
		},
		{
			name: "with prompt info",
			obs: &clients.ObservationDetail{
				ObservationInfo: clients.ObservationInfo{
					ID:      "obs-789",
					TraceID: "trace-abc",
					Type:    "GENERATION",
					Name:    "test",
				},
				PromptName:    "my-prompt",
				PromptVersion: &promptVersion,
			},
			want: "ID:           obs-789\n" +
				"Trace ID:     trace-abc\n" +
				"Type:         GENERATION\n" +
				"Name:         test\n" +
				"Start Time:   \n" +
				"End Time:     \n" +
				"\n--- Model ---\n" +
				"Model:   \n" +
				"\n--- Metrics ---\n" +
				"Latency:         -\n" +
				"Input Tokens:    -\n" +
				"Output Tokens:   -\n" +
				"Total Tokens:    -\n" +
				"Total Cost:      -\n" +
				"\n--- Prompt ---\n" +
				"Prompt Name:      my-prompt\n" +
				"Prompt Version:   3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintObservationDetail(&buf, tt.obs)
			if err != nil {
				t.Fatalf("PrintObservationDetail() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintStandaloneObservationList(t *testing.T) {
	cost := 0.0081
	latency := 9321.0
	tokens := 7474

	tests := []struct {
		name         string
		observations []clients.StandaloneObservationInfo
		meta         clients.CursorMeta
		columns      []string
		want         string
	}{
		{
			name:         "empty list prints message",
			observations: nil,
			columns:      inputs.DefaultStandaloneObservationColumns,
			want:         "No observations found.\n",
		},
		{
			name: "default columns with cursor",
			observations: []clients.StandaloneObservationInfo{
				{
					ID:          "obs-123",
					TraceID:     "trace-456",
					Type:        "GENERATION",
					Name:        "ChatGPT",
					Model:       "gpt-4",
					LatencyMs:   &latency,
					TotalCost:   &cost,
					TotalTokens: &tokens,
				},
			},
			meta:    clients.CursorMeta{NextCursor: "cursor-2"},
			columns: inputs.DefaultStandaloneObservationColumns,
			want: "ID        TRACE ID    TYPE         NAME      MODEL   LATENCY (ms)   COST        TOTAL TOKENS\n" +
				"obs-123   trace-456   GENERATION   ChatGPT   gpt-4   9321.00ms      $0.008100   7474\n" +
				"\nNext cursor: cursor-2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintStandaloneObservationList(&buf, tt.observations, tt.meta, tt.columns)
			if err != nil {
				t.Fatalf("PrintStandaloneObservationList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
