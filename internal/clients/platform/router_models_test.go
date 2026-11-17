package platform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
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

func TestAPIClientListRouterModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/models/router-models" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("projectId") != "proj-1" {
			t.Fatalf("projectId = %q", query.Get("projectId"))
		}
		if query.Get("page") != "2" {
			t.Fatalf("page = %q", query.Get("page"))
		}
		if query.Get("limit") != "10" {
			t.Fatalf("limit = %q", query.Get("limit"))
		}
		if query.Get("search") != "gpt" {
			t.Fatalf("search = %q", query.Get("search"))
		}
		if query.Get("region") != "us" {
			t.Fatalf("region = %q", query.Get("region"))
		}
		_, _ = io.WriteString(
			w,
			`{"models":[{"id":"m-1","modelName":"gpt-4o","endpointProvider":"openai","region":"us"}],"totalCount":1}`,
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	models, meta, err := client.ListRouterModels(
		context.Background(),
		"proj-1",
		RouterModelListOptions{Page: 2, Limit: 10, Search: "gpt", Region: "us"},
	)
	if err != nil {
		t.Fatalf("ListRouterModels() error = %v", err)
	}
	if len(models) != 1 || models[0].ID != "m-1" {
		t.Fatalf("unexpected models: %#v", models)
	}
	// Page is 1-indexed (opts.Page 2 -> 3); TotalPages = ceil(1/10) = 1.
	if meta.Page != 3 || meta.TotalItems != 1 || meta.TotalPages != 1 {
		t.Fatalf("unexpected meta: %#v", meta)
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
		Latency          *RouterModelLatency
	}
	azure := "azure"
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
	latency := &RouterModelLatency{
		Date:                    "2026-11-16",
		TimeToFirstTokenSeconds: floatPtr(3.628168674698795),
		SecondsPerOutputToken:   floatPtr(0.021332107236064685),
	}

	tests := []struct {
		name    string
		payload []byte
		want    decoded
	}{
		{
			name:    "backend without developer, uptime and latency",
			payload: read("by_id_before_developer.json"),
			want: decoded{
				EndpointProvider: "openai",
				Providers:        6,
				FirstProvider:    firstProvider(nil),
			},
		},
		{
			name:    "backend with developer, uptime and latency",
			payload: read("by_id.json"),
			want: decoded{
				Developer:        "openai",
				EndpointProvider: "openai",
				Providers:        6,
				FirstProvider:    firstProvider(floatPtr(92.01388888888889)),
				Latency:          latency,
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
				Latency:          latency,
			},
		},
		{
			name:    "null providers decode to none",
			payload: withField(read("by_id.json"), "providers", "null"),
			want:    decoded{Developer: "openai", EndpointProvider: "openai", Latency: latency},
		},
		{
			name:    "null latency decodes to nil",
			payload: withField(read("by_id.json"), "latency", "null"),
			want: decoded{
				Developer:        "openai",
				EndpointProvider: "openai",
				Providers:        6,
				FirstProvider:    firstProvider(floatPtr(92.01388888888889)),
			},
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
				Latency:          detail.Latency,
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
