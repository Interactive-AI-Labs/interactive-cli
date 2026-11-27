package platform

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func newEvalTestClient(t *testing.T, baseURL string) *APIClient {
	t.Helper()
	client, err := NewAPIClient(baseURL, 5*time.Second, "fake-token", "", nil)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}
	return client
}

func newListTestServer(
	t *testing.T,
	wantPath string,
	wantQuery url.Values,
	body string,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != wantPath {
			t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
		}
		if got := r.URL.Query(); !reflect.DeepEqual(got, wantQuery) {
			t.Errorf("query = %v, want %v", got, wantQuery)
		}
		_, _ = io.WriteString(w, body)
	}))
}

func TestDoAndReadIncludesStatusOnBodyError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{"success response", http.StatusCreated, "HTTP 201"},
		{"error response", http.StatusFailedDependency, "HTTP 424"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Length", "10")
					w.WriteHeader(tt.status)
					_, _ = io.WriteString(w, "short")
				}),
			)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			req, err := client.newRequest(context.Background(), http.MethodGet, "/")
			if err != nil {
				t.Fatalf("newRequest() error = %v", err)
			}
			_, err = client.doAndRead(req, "test")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("doAndRead() error = %v, want status %q", err, tt.want)
			}
		})
	}
}

func TestAPIClientListAnnotationQueues(t *testing.T) {
	const path = "/api/platform/v1/organizations/org-1/projects/proj-1/annotation-queues"
	const body = `{"success":true,"data":{"queues":[{"id":"q-1","name":"queue",` +
		`"count_completed_items":2,"count_pending_items":3}],` +
		`"meta":{"page":1,"total_pages":1,"total_items":1}}}`

	tests := []struct {
		name      string
		opts      AnnotationQueueListOptions
		wantQuery url.Values
	}{
		{
			name:      "omits sorting when unset",
			opts:      AnnotationQueueListOptions{Page: 1},
			wantQuery: url.Values{"page": {"1"}},
		},
		{
			name: "sends sort_by alone",
			opts: AnnotationQueueListOptions{Page: 1, SortBy: "name"},
			wantQuery: url.Values{
				"page":    {"1"},
				"sort_by": {"name"},
			},
		},
		{
			name: "sends sort_by and sort_order",
			opts: AnnotationQueueListOptions{Page: 1, SortBy: "name", SortOrder: "asc"},
			wantQuery: url.Values{
				"page":       {"1"},
				"sort_by":    {"name"},
				"sort_order": {"asc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newListTestServer(t, path, tt.wantQuery, body)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			queues, meta, _, err := client.ListAnnotationQueues(
				context.Background(),
				"org-1",
				"proj-1",
				tt.opts,
			)
			if err != nil {
				t.Fatalf("ListAnnotationQueues() error = %v", err)
			}
			if len(queues) != 1 || queues[0].ID != "q-1" {
				t.Fatalf("unexpected queues: %#v", queues)
			}
			if queues[0].CountCompletedItems != 2 || queues[0].CountPendingItems != 3 {
				t.Fatalf("unexpected item counts: %#v", queues[0])
			}
			if meta.TotalItems != 1 {
				t.Fatalf("unexpected meta: %#v", meta)
			}
		})
	}
}

func TestAPIClientListQueueItems(t *testing.T) {
	const path = "/api/platform/v1/organizations/org-1/projects/proj-1/" +
		"annotation-queues/queue-1/items"
	const body = `{"success":true,"data":{"items":[{"id":"i-1","created_at":"2025-01-01"}],` +
		`"meta":{"page":1,"total_pages":1,"total_items":1}}}`

	tests := []struct {
		name      string
		opts      QueueItemListOptions
		wantQuery url.Values
	}{
		{
			name:      "omits sorting and status when unset",
			opts:      QueueItemListOptions{Page: 1},
			wantQuery: url.Values{"page": {"1"}},
		},
		{
			name: "sends sort_by alone",
			opts: QueueItemListOptions{Page: 1, SortBy: "status"},
			wantQuery: url.Values{
				"page":    {"1"},
				"sort_by": {"status"},
			},
		},
		{
			name: "sends status alongside sorting",
			opts: QueueItemListOptions{
				Page:      1,
				Status:    "PENDING",
				SortBy:    "completed_at",
				SortOrder: "desc",
			},
			wantQuery: url.Values{
				"page":       {"1"},
				"status":     {"PENDING"},
				"sort_by":    {"completed_at"},
				"sort_order": {"desc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newListTestServer(t, path, tt.wantQuery, body)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			items, _, _, err := client.ListQueueItems(
				context.Background(),
				"org-1",
				"proj-1",
				"queue-1",
				tt.opts,
			)
			if err != nil {
				t.Fatalf("ListQueueItems() error = %v", err)
			}
			if len(items) != 1 || items[0].CreatedAt != "2025-01-01" {
				t.Fatalf("unexpected items: %#v", items)
			}
		})
	}
}

func TestAPIClientDeleteQueueResources(t *testing.T) {
	const base = "/api/platform/v1/organizations/org-1/projects/proj-1/annotation-queues"

	tests := []struct {
		name        string
		wantPath    string
		body        string
		wantMessage string
		call        func(*APIClient) (string, error)
	}{
		{
			name:        "delete annotation queue",
			wantPath:    base + "/q-1",
			body:        `{"success":true,"data":{"message":"Annotation queue deleted"}}`,
			wantMessage: "Annotation queue deleted",
			call: func(c *APIClient) (string, error) {
				return c.DeleteAnnotationQueue(context.Background(), "org-1", "proj-1", "q-1")
			},
		},
		{
			name:        "delete queue item without an API key",
			wantPath:    base + "/q-1/items/i-1",
			body:        `{"success":true,"data":{"message":"Annotation queue item deleted"}}`,
			wantMessage: "Annotation queue item deleted",
			call: func(c *APIClient) (string, error) {
				return c.DeleteQueueItem(context.Background(), "org-1", "proj-1", "q-1", "i-1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodDelete {
						t.Errorf("method = %s, want DELETE", r.Method)
					}
					if r.URL.Path != tt.wantPath {
						t.Errorf("path = %s, want %s", r.URL.Path, tt.wantPath)
					}
					_, _ = io.WriteString(w, tt.body)
				}),
			)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			if client.isApiKeyMode {
				t.Fatal("test client should be in bearer-token mode")
			}

			message, err := tt.call(client)
			if err != nil {
				t.Fatalf("delete error = %v", err)
			}
			if message != tt.wantMessage {
				t.Fatalf("message = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}

func TestAPIClientDeleteDatasetSendsTheNameAsAQuery(t *testing.T) {
	for _, dataset := range []string{"evals", "foo/runs/bar"} {
		t.Run(dataset, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					got := []any{r.Method, r.URL.Path, r.URL.Query()}
					want := []any{
						http.MethodDelete,
						"/api/platform/v1/organizations/org-1/projects/proj-1/datasets",
						url.Values{"dataset_name": {dataset}},
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("request = %v, want %v", got, want)
					}
					_, _ = io.WriteString(
						w,
						`{"success":true,"data":{"message":"Dataset deleted"}}`,
					)
				}),
			)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			if _, err := client.DeleteDataset(
				context.Background(),
				"org-1",
				"proj-1",
				dataset,
			); err != nil {
				t.Fatalf("DeleteDataset() error = %v", err)
			}
		})
	}
}

func TestListAgentDatasetRuns(t *testing.T) {
	const base = "/api/platform/v1/organizations/org-1/projects/proj-1/datasets"
	bodies := map[string]string{
		base + "?agent=a&limit=100":  `{"success":true,"data":{"datasets":[{"name":"d1"},{"name":"d/2"}]}}`,
		base + "/d1/runs?limit=5":    `{"success":true,"data":{"runs":[{"name":"old","created_at":"2026-01-01"}]}}`,
		base + "/d%2F2/runs?limit=5": `{"success":true,"data":{"runs":[{"name":"new","created_at":"2026-02-01"}]}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, bodies[r.URL.EscapedPath()+"?"+r.URL.RawQuery])
	}))
	defer server.Close()

	_, meta, raw, err := newEvalTestClient(t, server.URL).ListAgentDatasetRuns(
		context.Background(), "org-1", "proj-1", "a", DatasetRunListOptions{Limit: 5})
	got := fmt.Sprintf("%s %d %v", raw, meta.TotalItems, err)
	want := `{"runs":[` +
		`{"id":"","name":"new","description":"","dataset_id":"","dataset_name":"","metadata":null,"created_at":"2026-02-01","updated_at":""},` +
		`{"id":"","name":"old","description":"","dataset_id":"","dataset_name":"","metadata":null,"created_at":"2026-01-01","updated_at":""}` +
		`]} 2 <nil>`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
