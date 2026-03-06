package output

import (
	"bytes"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintTraceList(t *testing.T) {
	latency := 1.5
	cost := 0.012345

	tests := []struct {
		name    string
		traces  []clients.TraceInfo
		meta    clients.TraceMeta
		columns []string
		want    string
	}{
		{
			name:    "empty list prints message",
			traces:  []clients.TraceInfo{},
			meta:    clients.TraceMeta{},
			columns: inputs.DefaultTraceColumns,
			want:    "No traces found.\n",
		},
		{
			name:    "nil list prints message",
			traces:  nil,
			meta:    clients.TraceMeta{},
			columns: inputs.DefaultTraceColumns,
			want:    "No traces found.\n",
		},
		{
			name: "default columns",
			traces: []clients.TraceInfo{
				{
					ID:        "abc123",
					Name:      "my-trace",
					Timestamp: "2025-01-01",
					Latency:   &latency,
					TotalCost: &cost,
					Tags:      []string{"tag1"},
				},
			},
			meta:    clients.TraceMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns: inputs.DefaultTraceColumns,
			want: "ID       NAME       TIMESTAMP    LATENCY   COST        TAGS\n" +
				"abc123   my-trace   2025-01-01   1.50s     $0.012345   tag1\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "custom columns subset",
			traces: []clients.TraceInfo{
				{
					ID:   "abc123",
					Name: "my-trace",
				},
			},
			meta:    clients.TraceMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns: []string{"id", "name"},
			want: "ID       NAME\n" +
				"abc123   my-trace\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "nil latency and cost show dash",
			traces: []clients.TraceInfo{
				{
					ID:        "abc123",
					Name:      "test",
					Timestamp: "2025-01-01",
				},
			},
			meta:    clients.TraceMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns: []string{"id", "latency", "cost"},
			want: "ID       LATENCY   COST\n" +
				"abc123   -         -\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "tags truncated beyond 3",
			traces: []clients.TraceInfo{
				{
					ID:   "abc123",
					Tags: []string{"a", "b", "c", "d", "e"},
				},
			},
			meta:    clients.TraceMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns: []string{"id", "tags"},
			want: "ID       TAGS\n" +
				"abc123   a, b, c (+2 more)\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "session_id and user_id columns",
			traces: []clients.TraceInfo{
				{
					ID:        "abc123",
					UserID:    "user1",
					SessionID: "sess1",
				},
			},
			meta:    clients.TraceMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			columns: []string{"id", "user_id", "session_id"},
			want: "ID       USER ID   SESSION ID\n" +
				"abc123   user1     sess1\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "pagination info",
			traces: []clients.TraceInfo{
				{ID: "abc123"},
			},
			meta:    clients.TraceMeta{Page: 2, TotalPages: 5, TotalItems: 50},
			columns: []string{"id"},
			want: "ID\n" +
				"abc123\n" +
				"\nPage 2 of 5 (50 total items)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintTraceList(&buf, tt.traces, tt.meta, tt.columns)
			if err != nil {
				t.Fatalf("PrintTraceList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintTraceDetail(t *testing.T) {
	latency := 2.5
	cost := 0.05

	var buf bytes.Buffer
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			ID:          "abc123",
			Name:        "my-trace",
			Timestamp:   "2025-01-01",
			SessionID:   "sess1",
			UserID:      "user1",
			Environment: "production",
			Release:     "v1.0",
			Version:     "1",
			Public:      true,
			Latency:     &latency,
			TotalCost:   &cost,
			Tags:        []string{"tag1", "tag2"},
			HtmlPath:    "/project/traces/abc123",
		},
	}

	err := PrintTraceDetail(&buf, trace)
	if err != nil {
		t.Fatalf("PrintTraceDetail() error = %v", err)
	}

	got := buf.String()

	expected := []string{
		"ID:          abc123",
		"Name:        my-trace",
		"Session ID:  sess1",
		"User ID:     user1",
		"Environment: production",
		"Release:     v1.0",
		"Version:     1",
		"Public:      true",
		"Latency:     2.50s",
		"Total Cost:  $0.050000",
		"Tags:        tag1, tag2",
		"URL Path:    /project/traces/abc123",
	}

	for _, s := range expected {
		if !bytes.Contains([]byte(got), []byte(s)) {
			t.Errorf("output missing %q\ngot:\n%s", s, got)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	v := 1.5
	tests := []struct {
		name   string
		val    *float64
		suffix string
		want   string
	}{
		{"nil returns dash", nil, "s", "-"},
		{"value with suffix", &v, "s", "1.50s"},
		{"value with empty suffix", &v, "", "1.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFloat(tt.val, tt.suffix); got != tt.want {
				t.Errorf("formatFloat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCost(t *testing.T) {
	v := 0.012345
	tests := []struct {
		name string
		val  *float64
		want string
	}{
		{"nil returns dash", nil, "-"},
		{"value formatted with dollar sign", &v, "$0.012345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCost(tt.val); got != tt.want {
				t.Errorf("formatCost() = %q, want %q", got, tt.want)
			}
		})
	}
}
