package output

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
					ID:               "m-1",
					ModelName:        "gpt-4o",
					EndpointProvider: "openai",
					ContextLength:    intPtr(128000),
					Region:           "us",
				},
			},
			meta: platform.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			want: "NAME     DEVELOPER   PROVIDERS   CONTEXT   REGION   RECOMMENDED   ID\n" +
				"gpt-4o   openai      -           128k      Global   -             m-1\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "no developer under either name and empty provider names print dashes",
			models: []platform.RouterModel{
				{ID: "m-3", ModelName: "gpt-4o", Region: "us", ProviderNames: []string{}},
			},
			meta: platform.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			want: "NAME     DEVELOPER   PROVIDERS   CONTEXT   REGION   RECOMMENDED   ID\n" +
				"gpt-4o   -           -           -         Global   -             m-3\n" +
				"\nPage 1 of 1 (1 total items)\n",
		},
		{
			name: "recommendation ranks name their category, sorted",
			models: []platform.RouterModel{
				{
					ID:               "m-2",
					ModelName:        "claude-opus-5",
					ContextLength:    intPtr(200000),
					Developer:        "Anthropic",
					EndpointProvider: "anthropic",
					ProviderNames:    []string{"Amazon Bedrock", "Anthropic"},
					Region:           "eu",
					Recommendations:  map[string]int{"vision": 3, "chat": 1},
				},
			},
			meta: platform.PageMeta{Page: 1, TotalPages: 1, TotalItems: 1},
			want: "NAME            DEVELOPER   PROVIDERS   CONTEXT   REGION   RECOMMENDED          ID\n" +
				"claude-opus-5   Anthropic   2           200k      EU       chat #1, vision #3   m-2\n" +
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
						Name: "Azure",
						Tag:  &azure,
						Prices: map[string]float64{
							"input":            1e-7,
							"output":           5e-7,
							"input_cache_read": 1e-8,
						},
						ContextLength:     intPtr(1050000),
						MaxOutputTokens:   intPtr(128000),
						UptimeLastDay:     floatPtr(99.73),
						LatencyLast30m:    &platform.RouterModelPercentiles{P50: floatPtr(1043.5)},
						ThroughputLast30m: &platform.RouterModelPercentiles{P50: floatPtr(44.5)},
					},
					{
						Name:           "OpenAI",
						Tag:            &openai,
						Prices:         map[string]float64{"input": 0},
						UptimeLastDay:  floatPtr(0),
						LatencyLast30m: &platform.RouterModelPercentiles{},
					},
					{Name: "Amazon Bedrock"},
				},
			},
			want: "ID:             m-1\n" +
				"Model Name:     interactive/openai/gpt-6-luna\n" +
				"Marketing:      GPT-6 Luna\n" +
				"Developer:      OpenAI\n" +
				"Region:         Global (also in EU)\n" +
				"Recommended:    chat #3, vision #3\n" +
				"Context:        1.1M\n" +
				"Capabilities:   text, vision\n" +
				"List price (per 1M tokens):\n" +
				"  input:    $0.10\n" +
				"  output:   $0.50\n" +
				"\n" +
				"PROVIDER         TAG      INPUT ($/1M)   OUTPUT ($/1M)   CACHE READ ($/1M)   CONTEXT   MAX OUTPUT   UPTIME (1D)   LATENCY (P50 30M)   TOK/S (P50 30M)\n" +
				"Azure            azure    $0.10          $0.50           $0.01               1.1M      128k         99.73%        1.04s               45\n" +
				"OpenAI           openai   $0.00          -               -                   -         -            0.00%         -                   -\n" +
				"Amazon Bedrock   -        -              -               -                   -         -            -             -                   -\n",
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
		{"token price per 1M", prices, "input", "$0.10"},
		{"zero price", prices, "output", "$0.00"},
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

func TestRouterModelPriceText(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		perToken float64
		want     string
	}{
		{"input", "input", 0.000003, "$3.00"},
		{"output", "output", 0.000015, "$15.00"},
		{"cache read", "input_cache_read", 1.014e-9, "$0.001014"},
		{"cache creation", "input_cache_creation", 0.00000375, "$3.75"},
		{"reasoning", "output_reasoning", 2.028e-7, "$0.2028"},
		{"per query", "search_units", 0.01, "$0.01 per query"},
		{"per 1M characters", "input_characters", 0.00008, "$80.00 per 1M characters"},
		{"per hour", "input_seconds", 0.0000611111111111, "$0.22 per hour"},
		{"per hour, cheaper price", "input_seconds", 0.0000055555555555, "$0.02 per hour"},
		{"unknown key images", "input_images", 0.04, "0.04"},
		{"unknown key audio", "input_audio", 1e-7, "0.0000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routerModelPriceText(tt.key, tt.perToken); got != tt.want {
				t.Errorf(
					"routerModelPriceText(%q, %v) = %q, want %q",
					tt.key,
					tt.perToken,
					got,
					tt.want,
				)
			}
		})
	}
}

func readRouterModelBody(t *testing.T, name string, into any) {
	t.Helper()
	data, err := os.ReadFile(
		filepath.Join("..", "clients", "platform", "testdata", "router_models", name),
	)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if err := json.Unmarshal(data, into); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

// Lines are compared with runs of spaces collapsed, so column widths set by other rows do not matter.
func TestPrintRouterModelDetailOnDevBodies(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		edit    func(*platform.RouterModelDetail)
		want    []string
		absent  []string
	}{
		{
			name:    "sub-cent prices keep their digits per 1M",
			fixture: "by_id_deepseek_before_latency_dev.json",
			want: []string{
				"List price (per 1M tokens):",
				"input: $0.0507",
				"input_cache_read: $0.001014",
				"output: $0.2028",
				"output_reasoning: $0.2028",
				"Baidu baidu/fp8 $0.0507 $0.2028 $0.001014 1.0M 393k 100.00% - -",
			},
			absent: []string{"List price:"},
		},
		{
			name:    "round prices print two decimals, no float noise",
			fixture: "by_id_sonnet_dev.json",
			want: []string{
				"Context: 1.0M",
				"List price (per 1M tokens):",
				"input: $3.00",
				"input_cache_creation: $3.75",
				"input_cache_read: $0.30",
				"output: $15.00",
				"output_reasoning: $15.00",
				"PROVIDER TAG INPUT ($/1M) OUTPUT ($/1M) CACHE READ ($/1M) CONTEXT MAX OUTPUT UPTIME (1D) LATENCY (P50 30M) TOK/S (P50 30M)",
				"Anthropic anthropic $3.00 $15.00 $0.30 1.0M 64k 100.00% 1.03s 46",
				"Azure azure/global $3.00 $15.00 $0.30 200k 64k - 1.57s 43",
			},
			absent: []string{"List price:"},
		},
		{
			name:    "a unit price keeps its unit",
			fixture: "by_id_eleven_before_latency_dev.json",
			want: []string{
				"List price:",
				"input_characters: $80.00 per 1M characters",
				"ElevenLabs elevenlabs - - - - - 99.74% - -",
			},
			absent: []string{"List price (per 1M tokens):"},
		},
		{
			name:    "an unknown price key is printed as sent, under the model's other prices",
			fixture: "by_id_sonnet_dev.json",
			edit: func(m *platform.RouterModelDetail) {
				m.Prices["input_images"] = 0.04
			},
			want: []string{
				"List price (per 1M tokens):",
				"output_reasoning: $15.00",
				"List price:",
				"input_images: 0.04",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m platform.RouterModelDetail
			readRouterModelBody(t, tt.fixture, &m)
			if tt.edit != nil {
				tt.edit(&m)
			}
			var buf bytes.Buffer
			if err := PrintRouterModelDetail(&buf, &m); err != nil {
				t.Fatalf("PrintRouterModelDetail() error = %v", err)
			}
			var lines []string
			for _, l := range strings.Split(buf.String(), "\n") {
				lines = append(lines, strings.Join(strings.Fields(l), " "))
			}
			next := 0
			for _, w := range tt.want {
				i := slices.Index(lines[next:], w)
				if i < 0 {
					t.Fatalf("line %q not found in order in:\n%s", w, buf.String())
				}
				next += i + 1
			}
			for _, a := range tt.absent {
				if slices.Contains(lines, a) {
					t.Errorf("line %q should be absent in:\n%s", a, buf.String())
				}
			}
		})
	}
}
