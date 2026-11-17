package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func intPtr(i int) *int { return &i }

func floatPtr(f float64) *float64 { return &f }

func TestPrintRouterModelList(t *testing.T) {
	tests := []struct {
		name   string
		models []platform.RouterModel
		meta   platform.PageMeta
		want   string
	}{
		{
			name:   "empty list prints message",
			models: nil,
			want:   "No models found.\n",
		},
		{
			name: "single model",
			models: []platform.RouterModel{
				{
					ID:            "m-1",
					ModelName:     "gpt-4o",
					ContextLength: intPtr(128000),
					Region:        "us",
				},
			},
			meta: platform.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			want: "NAME     CONTEXT   REGION   RECOMMENDED   ID\n" +
				"gpt-4o   128000    us       -             m-1\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "recommendation ranks name their category, sorted",
			models: []platform.RouterModel{
				{
					ID:              "m-2",
					ModelName:       "claude-opus-5",
					ContextLength:   intPtr(200000),
					Region:          "eu",
					Recommendations: map[string]int{"vision": 3, "chat": 1},
				},
			},
			meta: platform.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			want: "NAME            CONTEXT   REGION   RECOMMENDED          ID\n" +
				"claude-opus-5   200000    eu       chat #1, vision #3   m-2\n" +
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
	azure, openai := "azure", "openai"
	tests := []struct {
		name  string
		model platform.RouterModelDetail
		want  string
	}{
		{
			name: "model with provider rows",
			model: platform.RouterModelDetail{
				RouterModel: platform.RouterModel{
					ID:               "m-1",
					ModelName:        "interactive/openai/gpt-6-luna",
					MarketingName:    "GPT-6 Luna",
					Developer:        "OpenAI",
					EndpointProvider: "openai",
					Region:           "us",
					HasOtherRegion:   true,
					ContextLength:    intPtr(1050000),
					Capabilities:     []string{"text", "vision"},
					Prices:           map[string]float64{"input": 1e-7, "output": 5e-7},
					Recommendations:  map[string]int{"chat": 3, "vision": 3},
				},
				Providers: []platform.RouterModelProvider{
					{
						Name:            "Azure",
						Tag:             &azure,
						Prices:          map[string]float64{"input": 1e-7, "output": 5e-7},
						ContextLength:   intPtr(1050000),
						MaxOutputTokens: intPtr(128000),
						UptimeLastDay:   floatPtr(99.73),
					},
					{
						Name:          "OpenAI",
						Tag:           &openai,
						Prices:        map[string]float64{"input": 0},
						UptimeLastDay: floatPtr(0),
					},
					{Name: "Amazon Bedrock"},
				},
				Latency: &platform.RouterModelLatency{
					Date:                    "2026-11-16",
					TimeToFirstTokenSeconds: floatPtr(2.7),
					SecondsPerOutputToken:   floatPtr(0.033),
				},
			},
			want: "ID:             m-1\n" +
				"Model Name:     interactive/openai/gpt-6-luna\n" +
				"Marketing:      GPT-6 Luna\n" +
				"Developer:      OpenAI\n" +
				"Region:         Global (also in EU)\n" +
				"Recommended:    chat #3, vision #3\n" +
				"Context:        1050000\n" +
				"Latency:        2.7 s to first token, 0.033 s per output token (2026-11-16, UTC)\n" +
				"Capabilities:   text, vision\n" +
				"Prices:\n" +
				"  input:    0.0000001\n" +
				"  output:   0.0000005\n" +
				"\n" +
				"PROVIDER         TAG      INPUT       OUTPUT      CONTEXT   MAX OUTPUT   UPTIME (1D)\n" +
				"Azure            azure    0.0000001   0.0000005   1050000   128000       99.73%\n" +
				"OpenAI           openai   0           -           -         -            0.00%\n" +
				"Amazon Bedrock   -        -           -           -         -            -\n",
		},
		{
			name: "empty recommendations, context and prices drop their lines; no rows says none",
			model: platform.RouterModelDetail{
				RouterModel: platform.RouterModel{
					ID:               "m-2",
					ModelName:        "interactive/google/gemini-2.5-flash",
					EndpointProvider: "google",
					Region:           "eu",
				},
			},
			want: "ID:           m-2\n" +
				"Model Name:   interactive/google/gemini-2.5-flash\n" +
				"Developer:    google\n" +
				"Region:       EU\n" +
				"Latency:      -\n" +
				"Providers:    none\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintRouterModelDetail(&buf, &tt.model); err != nil {
				t.Fatalf("PrintRouterModelDetail() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestRouterModelRegionLine(t *testing.T) {
	tests := []struct {
		name           string
		region         string
		hasOtherRegion bool
		want           string
	}{
		{"global only", "us", false, "Global"},
		{"global, also in eu", "us", true, "Global (also in EU)"},
		{"eu only", "eu", false, "EU"},
		{"eu, also global", "eu", true, "EU (also in Global)"},
		{"unknown region is printed as sent", "apac", true, "apac"},
		{"no region", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routerModelRegionLine(tt.region, tt.hasOtherRegion); got != tt.want {
				t.Errorf("routerModelRegionLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRouterModelPrice(t *testing.T) {
	prices := map[string]float64{"input": 1e-07, "output": 0}
	tests := []struct {
		name   string
		prices map[string]float64
		key    string
		want   string
	}{
		{"small price keeps its digits", prices, "input", "0.0000001"},
		{"zero price", prices, "output", "0"},
		{"missing key", prices, "input_cache_read", "-"},
		{"no prices", nil, "input", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routerModelPrice(tt.prices, tt.key); got != tt.want {
				t.Errorf("routerModelPrice() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRouterModelLatency(t *testing.T) {
	tests := []struct {
		name    string
		latency *platform.RouterModelLatency
		want    string
	}{
		{"no latency", nil, "-"},
		{
			"both figures",
			&platform.RouterModelLatency{
				Date:                    "2026-11-16",
				TimeToFirstTokenSeconds: floatPtr(3.628168674698795),
				SecondsPerOutputToken:   floatPtr(0.021332107236064685),
			},
			"3.63 s to first token, 0.0213 s per output token (2026-11-16, UTC)",
		},
		{
			"no time to first token",
			&platform.RouterModelLatency{
				Date:                  "2026-11-16",
				SecondsPerOutputToken: floatPtr(0.033),
			},
			"- to first token, 0.033 s per output token (2026-11-16, UTC)",
		},
		{
			"no seconds per output token",
			&platform.RouterModelLatency{Date: "2026-11-16", TimeToFirstTokenSeconds: floatPtr(0)},
			"0 s to first token, - per output token (2026-11-16, UTC)",
		},
		{"neither figure", &platform.RouterModelLatency{Date: "2026-11-16"}, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routerModelLatency(tt.latency); got != tt.want {
				t.Errorf("routerModelLatency() = %q, want %q", got, tt.want)
			}
		})
	}
}
