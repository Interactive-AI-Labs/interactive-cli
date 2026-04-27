package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintTraceList(t *testing.T) {
	latencyMs := 1500.0
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
					LatencyMs: &latencyMs,
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
	latencyMs := 2500.0
	cost := 0.05

	tests := []struct {
		name  string
		trace *clients.TraceDetail
		want  string
	}{
		{
			name: "all fields populated",
			trace: &clients.TraceDetail{
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
					LatencyMs:   &latencyMs,
					TotalCost:   &cost,
					Tags:        []string{"tag1", "tag2"},
					HtmlPath:    "/project/traces/abc123",
				},
				Input:    []byte(`{"role":"user"}`),
				Output:   []byte(`"hello"`),
				Metadata: []byte(`{"key":"val"}`),
			},
			want: "ID:            abc123\n" +
				"Name:          my-trace\n" +
				"Timestamp:     2025-01-01\n" +
				"Session ID:    sess1\n" +
				"User ID:       user1\n" +
				"Environment:   production\n" +
				"Release:       v1.0\n" +
				"Version:       1\n" +
				"Public:        true\n" +
				"\n--- Metrics ---\n" +
				"Latency:             2.50s\n" +
				"Total Cost:          $0.050000\n" +
				"Observation Count:   -\n" +
				"Input Tokens:        -\n" +
				"Output Tokens:       -\n" +
				"Total Tokens:        -\n" +
				"\nTags:       tag1, tag2\n" +
				"URL Path:   /project/traces/abc123\n" +
				"\nInput:\n  {\n    \"role\": \"user\"\n  }\n" +
				"\nOutput:\n  \"hello\"\n" +
				"\nMetadata:\n  {\n    \"key\": \"val\"\n  }\n",
		},
		{
			name: "minimal fields",
			trace: &clients.TraceDetail{
				TraceInfo: clients.TraceInfo{
					ID:        "def456",
					Name:      "minimal",
					Timestamp: "2025-06-01",
				},
			},
			want: "ID:            def456\n" +
				"Name:          minimal\n" +
				"Timestamp:     2025-06-01\n" +
				"Session ID:    \n" +
				"User ID:       \n" +
				"Environment:   \n" +
				"Release:       \n" +
				"Version:       \n" +
				"Public:        false\n" +
				"\n--- Metrics ---\n" +
				"Latency:             -\n" +
				"Total Cost:          -\n" +
				"Observation Count:   -\n" +
				"Input Tokens:        -\n" +
				"Output Tokens:       -\n" +
				"Total Tokens:        -\n",
		},
		{
			name: "null json fields are hidden",
			trace: &clients.TraceDetail{
				TraceInfo: clients.TraceInfo{
					ID:   "ghi789",
					Name: "null-json",
				},
				Input:    []byte(`null`),
				Output:   []byte(`null`),
				Metadata: []byte(`null`),
			},
			want: "ID:            ghi789\n" +
				"Name:          null-json\n" +
				"Timestamp:     \n" +
				"Session ID:    \n" +
				"User ID:       \n" +
				"Environment:   \n" +
				"Release:       \n" +
				"Version:       \n" +
				"Public:        false\n" +
				"\n--- Metrics ---\n" +
				"Latency:             -\n" +
				"Total Cost:          -\n" +
				"Observation Count:   -\n" +
				"Input Tokens:        -\n" +
				"Output Tokens:       -\n" +
				"Total Tokens:        -\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintTraceDetail(&buf, tt.trace)
			if err != nil {
				t.Fatalf("PrintTraceDetail() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrettyJSON(t *testing.T) {
	tests := []struct {
		name     string
		raw      json.RawMessage
		unescape bool
		want     string
	}{
		{
			name:     "object",
			raw:      json.RawMessage(`{"key":"val"}`),
			unescape: false,
			want:     "{\n  \"key\": \"val\"\n}",
		},
		{
			name:     "array",
			raw:      json.RawMessage(`[1,2,3]`),
			unescape: false,
			want:     "[\n  1,\n  2,\n  3\n]",
		},
		{
			name:     "string value",
			raw:      json.RawMessage(`"hello"`),
			unescape: false,
			want:     `"hello"`,
		},
		{
			name:     "number value",
			raw:      json.RawMessage(`42`),
			unescape: false,
			want:     "42",
		},
		{
			name:     "nested object indentation",
			raw:      json.RawMessage(`{"a":{"b":"c"}}`),
			unescape: false,
			want:     "{\n  \"a\": {\n    \"b\": \"c\"\n  }\n}",
		},
		{
			name:     "invalid json returns raw",
			raw:      json.RawMessage(`{not valid`),
			unescape: false,
			want:     "{not valid",
		},
		{
			name:     "unescape newlines and tabs",
			raw:      json.RawMessage(`"line1\nline2\tcol"`),
			unescape: true,
			want:     "\"line1\nline2\tcol\"",
		},
		{
			name:     "unescape backslashes",
			raw:      json.RawMessage(`"a\\\\b"`),
			unescape: true,
			want:     "\"a\\\\b\"",
		},
		{
			name:     "unescape quotes",
			raw:      json.RawMessage(`"say \\\"hello\\\""`),
			unescape: true,
			want:     "\"say \\\"hello\\\"\"",
		},
		{
			name:     "no unescape when false",
			raw:      json.RawMessage(`"line1\nline2"`),
			unescape: false,
			want:     `"line1\nline2"`,
		},
		{
			name:     "empty object",
			raw:      json.RawMessage(`{}`),
			unescape: false,
			want:     "{}",
		},
		{
			name:     "empty array",
			raw:      json.RawMessage(`[]`),
			unescape: false,
			want:     "[]",
		},
		{
			name:     "literal backslash-n in data is preserved",
			raw:      json.RawMessage(`{"input":"print(\"line1\\nline2\")"}`),
			unescape: true,
			want:     "{\n  \"input\": \"print(\"line1\\nline2\")\"\n}",
		},
		{
			name:     "multi-level escaping not collapsed",
			raw:      json.RawMessage(`"\\\\n"`),
			unescape: true,
			want:     "\"\\\\n\"",
		},
		{
			name:     "unicode escape decoded",
			raw:      json.RawMessage(`"hello\u0021"`),
			unescape: true,
			want:     "\"hello!\"",
		},
		{
			name:     "html chars not escaped",
			raw:      json.RawMessage(`{"reason":"A \u0026 B \u003c C"}`),
			unescape: false,
			want:     "{\n  \"reason\": \"A & B < C\"\n}",
		},
		{
			name:     "null value preserved",
			raw:      json.RawMessage(`{"locationPosition":null}`),
			unescape: false,
			want:     "{\n  \"locationPosition\": null\n}",
		},
		{
			name:     "boolean and numeric values",
			raw:      json.RawMessage(`{"ok":1,"sent":true,"approved":false}`),
			unescape: false,
			want:     "{\n  \"approved\": false,\n  \"ok\": 1,\n  \"sent\": true\n}",
		},
		{
			name:     "array of objects",
			raw:      json.RawMessage(`{"tags":["WRONG_ADDRESS","PROBLEMATIC_APPLICANT_DATA"]}`),
			unescape: false,
			want:     "{\n  \"tags\": [\n    \"WRONG_ADDRESS\",\n    \"PROBLEMATIC_APPLICANT_DATA\"\n  ]\n}",
		},
		{
			name:     "non-ascii characters preserved",
			raw:      json.RawMessage(`{"comment":"Café résumé naïve"}`),
			unescape: false,
			want:     "{\n  \"comment\": \"Café résumé naïve\"\n}",
		},
		{
			name: "nested objects with arrays",
			raw: json.RawMessage(
				`{"info":{"name":"Alice","items":[{"color":"red","size":"M"}]}}`,
			),
			unescape: false,
			want:     "{\n  \"info\": {\n    \"items\": [\n      {\n        \"color\": \"red\",\n        \"size\": \"M\"\n      }\n    ],\n    \"name\": \"Alice\"\n  }\n}",
		},
		{
			name:     "empty notes array",
			raw:      json.RawMessage(`{"notes":[],"type":"individual"}`),
			unescape: false,
			want:     "{\n  \"notes\": [],\n  \"type\": \"individual\"\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prettyJSON(tt.raw, tt.unescape); got != tt.want {
				t.Errorf("prettyJSON() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
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
