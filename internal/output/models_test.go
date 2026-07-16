package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func intPtr(i int) *int { return &i }

func TestPrintRouterModelList(t *testing.T) {
	tests := []struct {
		name   string
		models []clients.RouterModel
		meta   clients.PageMeta
		want   string
	}{
		{
			name:   "empty list prints message",
			models: nil,
			want:   "No models found.\n",
		},
		{
			name: "single model",
			models: []clients.RouterModel{
				{
					ID:            "m-1",
					ModelName:     "gpt-4o",
					ContextLength: intPtr(128000),
					Region:        "us",
				},
			},
			meta: clients.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			want: "NAME     CONTEXT   REGION   ID\n" +
				"gpt-4o   128000    us       m-1\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintRouterModelList(&buf, tt.models, tt.meta); err != nil {
				t.Fatalf("PrintRouterModelList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintRouterModelDetail(t *testing.T) {
	var buf bytes.Buffer
	model := &clients.RouterModel{
		ID:           "m-1",
		ModelName:    "gpt-4o",
		MatchPattern: "gpt-4o*",
		Region:       "us",
		Capabilities: []string{"text", "vision"},
		Prices:       map[string]float64{"input": 0.01, "output": 0.03},
	}
	if err := PrintRouterModelDetail(&buf, model); err != nil {
		t.Fatalf("PrintRouterModelDetail() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{
		"ID:             m-1",
		"Model Name:     gpt-4o",
		"Capabilities:   text, vision",
		"Prices:",
		"  input:    0.01",
		"  output:   0.03",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, got)
		}
	}
}
