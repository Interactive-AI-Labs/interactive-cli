package platform

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func floatPtr(f float64) *float64 { return &f }

func intPtr(i int) *int { return &i }

// readRouterModelFixture reads a response captured from the platform.
func readRouterModelFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "router_models", name))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	return data
}

func TestRouterModelUnmarshalRecommendations(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    map[string]int
		wantErr bool
	}{
		{
			name:    "ranks decode per category",
			payload: `{"id":"m-1","recommendations":{"chat":2,"vision":5}}`,
			want:    map[string]int{"chat": 2, "vision": 5},
		},
		{
			name:    "absent field leaves the map nil",
			payload: `{"id":"m-1"}`,
			want:    nil,
		},
		{
			name:    "null decodes to nil",
			payload: `{"id":"m-1","recommendations":null}`,
			want:    nil,
		},
		{
			name:    "empty object decodes to empty map",
			payload: `{"id":"m-1","recommendations":{}}`,
			want:    map[string]int{},
		},
		{
			name:    "non-numeric rank is rejected",
			payload: `{"id":"m-1","recommendations":{"chat":"top"}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got RouterModel
			err := json.Unmarshal([]byte(tt.payload), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %#v", got.Recommendations)
				}
				return
			}
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got.Recommendations, tt.want) {
				t.Errorf("Recommendations = %#v, want %#v", got.Recommendations, tt.want)
			}
		})
	}
}

func TestDecodeRouterModelList(t *testing.T) {
	type row struct {
		Developer        string
		EndpointProvider string
		ProviderNames    []string
	}
	tests := []struct {
		name      string
		body      []byte
		opts      RouterModelListOptions
		wantErr   bool
		wantMeta  PageMeta
		wantCount int
		wantFirst row
	}{
		{
			name:      "provider names; a partial page counts as a page",
			body:      readRouterModelFixture(t, "list.json"),
			opts:      RouterModelListOptions{Limit: 10},
			wantMeta:  PageMeta{Page: 1, Limit: 10, TotalItems: 1, TotalPages: 1},
			wantCount: 1,
			wantFirst: row{"openai", "openai", []string{"Azure", "OpenAI"}},
		},
		{
			name:      "null provider names stay nil; a full page adds no extra page",
			body:      readRouterModelFixture(t, "list_no_provider_rows.json"),
			opts:      RouterModelListOptions{Limit: 2},
			wantMeta:  PageMeta{Page: 1, Limit: 2, TotalItems: 2, TotalPages: 1},
			wantCount: 2,
			wantFirst: row{"google", "google", nil},
		},
		{
			name:      "backend without developer and provider names; no limit gives no page count",
			body:      readRouterModelFixture(t, "list_before_developer.json"),
			wantMeta:  PageMeta{Page: 1, TotalItems: 1},
			wantCount: 1,
			wantFirst: row{"", "openai", nil},
		},
		{
			name:     "page past the end reports the 1-indexed page and totalCount",
			body:     []byte(`{"models":[],"totalCount":7}`),
			opts:     RouterModelListOptions{Page: 2, Limit: 5},
			wantMeta: PageMeta{Page: 3, Limit: 5, TotalItems: 7, TotalPages: 2},
		},
		{
			name:    "missing models is an error",
			body:    []byte(`{"totalCount":0}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models, meta, raw, err := decodeRouterModelList(tt.body, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d models", len(models))
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeRouterModelList() error = %v", err)
			}
			if len(models) != tt.wantCount {
				t.Fatalf("len(models) = %d, want %d", len(models), tt.wantCount)
			}
			if meta != tt.wantMeta {
				t.Errorf("meta = %#v, want %#v", meta, tt.wantMeta)
			}
			if len(models) > 0 {
				got := row{models[0].Developer, models[0].EndpointProvider, models[0].ProviderNames}
				if !reflect.DeepEqual(got, tt.wantFirst) {
					t.Errorf("first model = %#v, want %#v", got, tt.wantFirst)
				}
			}
			if len(raw) == 0 || raw[0] != '[' || !bytes.Contains(tt.body, raw) {
				t.Errorf("raw = %.40s, want the models array as received", raw)
			}
		})
	}
}

func TestRouterModelDetailUnmarshal(t *testing.T) {
	read := func(name string) []byte { return readRouterModelFixture(t, name) }
	withField := func(payload []byte, key, value string) []byte {
		t.Helper()
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		fields[key] = json.RawMessage(value)
		edited, err := json.Marshal(fields)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		return edited
	}

	type decoded struct {
		Developer        string
		EndpointProvider string
		Providers        int
		FirstProvider    *RouterModelProvider
	}
	azure, anthropic := "azure", "anthropic"
	firstProvider := func(uptime *float64) *RouterModelProvider {
		return &RouterModelProvider{
			Name: "Azure",
			Tag:  &azure,
			Prices: map[string]float64{
				"input":                1e-7,
				"output":               5e-7,
				"input_cache_read":     1e-8,
				"input_cache_creation": 1.25e-7,
				"output_reasoning":     5e-7,
			},
			ContextLength:   intPtr(1050000),
			MaxOutputTokens: intPtr(128000),
			UptimeLastDay:   uptime,
		}
	}
	tests := []struct {
		name    string
		payload []byte
		want    decoded
	}{
		{
			name:    "backend without developer and uptime",
			payload: read("by_id_before_developer.json"),
			want: decoded{
				EndpointProvider: "openai",
				Providers:        6,
				FirstProvider:    firstProvider(nil),
			},
		},
		{
			name:    "backend with developer and uptime",
			payload: read("by_id.json"),
			want: decoded{
				Developer:        "openai",
				EndpointProvider: "openai",
				Providers:        6,
				FirstProvider:    firstProvider(floatPtr(92.01388888888889)),
			},
		},
		{
			name:    "developer is read from its own field",
			payload: withField(read("by_id.json"), "developer", `"OpenAI"`),
			want: decoded{
				Developer:        "OpenAI",
				EndpointProvider: "openai",
				Providers:        6,
				FirstProvider:    firstProvider(floatPtr(92.01388888888889)),
			},
		},
		{
			name:    "provider latency and throughput percentiles",
			payload: read("by_id_sonnet_dev.json"),
			want: decoded{
				Developer:        "anthropic",
				EndpointProvider: "anthropic",
				Providers:        4,
				FirstProvider: &RouterModelProvider{
					Name: "Anthropic",
					Tag:  &anthropic,
					Prices: map[string]float64{
						"input":                3e-6,
						"output":               0.000015,
						"input_cache_read":     3e-7,
						"input_cache_creation": 3.75e-6,
						"output_reasoning":     0.000015,
					},
					ContextLength:     intPtr(1000000),
					MaxOutputTokens:   intPtr(64000),
					UptimeLastDay:     floatPtr(100),
					LatencyLast30m:    &RouterModelPercentiles{P50: floatPtr(1028)},
					ThroughputLast30m: &RouterModelPercentiles{P50: floatPtr(46)},
				},
			},
		},
		{
			name:    "null providers decode to none",
			payload: withField(read("by_id.json"), "providers", "null"),
			want:    decoded{Developer: "openai", EndpointProvider: "openai"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var detail RouterModelDetail
			if err := json.Unmarshal(tt.payload, &detail); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			got := decoded{
				Developer:        detail.Developer,
				EndpointProvider: detail.EndpointProvider,
				Providers:        len(detail.Providers),
			}
			if len(detail.Providers) > 0 {
				got.FirstProvider = &detail.Providers[0]
			}
			if !reflect.DeepEqual(got, tt.want) {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				t.Errorf("decoded = %s\nwant      %s", gotJSON, wantJSON)
			}
		})
	}
}
