package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintStructuredYAML(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		contains []string
	}{
		{
			name: "normalizes raw message",
			value: struct {
				Name   string          `json:"name"`
				Prompt json.RawMessage `json:"prompt"`
			}{
				Name:   "routine",
				Prompt: json.RawMessage(`{"steps":["one","two"]}`),
			},
			contains: []string{"name: routine", "prompt:", "steps:", "- one", "- two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintStructuredYAML(&buf, tt.value); err != nil {
				t.Fatalf("PrintStructuredYAML() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("PrintStructuredYAML() missing %q in output:\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintRawYAML(t *testing.T) {
	tests := []struct {
		name     string
		raw      json.RawMessage
		contains []string
	}{
		{
			name:     "api envelope",
			raw:      json.RawMessage(`{"success":true,"data":{"traces":[{"id":"tr-1"}]}}`),
			contains: []string{"success: true", "data:", "traces:", "id: tr-1"},
		},
		{
			name:     "preserves numeric types",
			raw:      json.RawMessage(`{"count":14,"cost":17.5,"nested":{"limit":1}}`),
			contains: []string{"count: 14", "cost: 17.5", "limit: 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintRawYAML(&buf, tt.raw); err != nil {
				t.Fatalf("PrintRawYAML() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("PrintRawYAML() missing %q in output:\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintStructuredJSON(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name: "uses json tags",
			value: struct {
				TotalCount int `json:"totalCount"`
			}{TotalCount: 2},
			want: "{\n  \"totalCount\": 2\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintStructuredJSON(&buf, tt.value); err != nil {
				t.Fatalf("PrintStructuredJSON() error = %v", err)
			}

			if buf.String() != tt.want {
				t.Fatalf("PrintStructuredJSON() = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}
