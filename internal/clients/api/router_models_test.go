package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

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

func TestAPIClientGetRouterModelByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/models/router-models/by-id/m-1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("projectId") != "proj-1" {
			t.Fatalf("projectId = %q", r.URL.Query().Get("projectId"))
		}
		_, _ = io.WriteString(
			w,
			`{"id":"m-1","modelName":"gpt-4o","endpointProvider":"openai","region":"us","prices":{"input":0.01}}`,
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

	model, err := client.GetRouterModelByID(context.Background(), "proj-1", "m-1")
	if err != nil {
		t.Fatalf("GetRouterModelByID() error = %v", err)
	}
	if model.ID != "m-1" || model.ModelName != "gpt-4o" {
		t.Fatalf("unexpected model: %#v", model)
	}
	if model.Prices["input"] != 0.01 {
		t.Fatalf("unexpected prices: %#v", model.Prices)
	}
}
