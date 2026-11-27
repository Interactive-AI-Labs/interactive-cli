package platform

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
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

func TestNewestFirst(t *testing.T) {
	tests := []struct{ in, want string }{
		{"d1:01 d2:02", "d2:02 d1:01"},
		{"d1:03 d1:01 d2:02", "d1:03 d2:02 d1:01"},
		{"d1:01 d2:01", "d1:01 d2:01"},
	}
	for _, tt := range tests {
		var runs []DatasetRunInfo
		for _, f := range strings.Fields(tt.in) {
			dataset, created, _ := strings.Cut(f, ":")
			runs = append(runs, DatasetRunInfo{DatasetName: dataset, CreatedAt: created})
		}
		newestFirst(runs)
		got := make([]string, len(runs))
		for i, r := range runs {
			got[i] = r.DatasetName + ":" + r.CreatedAt
		}
		if strings.Join(got, " ") != tt.want {
			t.Errorf("newestFirst(%s) = %v, want %s", tt.in, got, tt.want)
		}
	}
}

func TestPageOf(t *testing.T) {
	newest := []int{1, 2, 3, 4}
	tests := []struct {
		total, page, limit int
		want               []int
		meta               PageMeta
	}{
		{4, 1, 2, []int{1, 2}, PageMeta{Page: 1, Limit: 2, TotalItems: 4, TotalPages: 2}},
		{9, 2, 2, []int{3, 4}, PageMeta{Page: 2, Limit: 2, TotalItems: 9, TotalPages: 5}},
		{4, 3, 2, []int{}, PageMeta{Page: 3, Limit: 2, TotalItems: 4, TotalPages: 2}},
		{
			4,
			math.MaxInt,
			50,
			[]int{},
			PageMeta{Page: math.MaxInt, Limit: 50, TotalItems: 4, TotalPages: 1},
		},
		{
			4,
			2,
			math.MaxInt,
			[]int{},
			PageMeta{Page: 2, Limit: math.MaxInt, TotalItems: 4, TotalPages: 1},
		},
	}
	for _, tt := range tests {
		got, meta := pageOf(newest, tt.total, tt.page, tt.limit)
		if !slices.Equal(got, tt.want) || meta != tt.meta {
			t.Errorf("pageOf(%d, %d, %d) = %v %+v, want %v %+v",
				tt.total, tt.page, tt.limit, got, meta, tt.want, tt.meta)
		}
	}
}

func TestReadPages(t *testing.T) {
	tests := []struct {
		name             string
		want, totalPages int
		got              string
	}{
		{"stops at TotalPages", math.MaxInt, 3, "3 1000 <nil>"},
		{"stops at want", 2, math.MaxInt, "2 1000 <nil>"},
		{"stops at maxPages", math.MaxInt, math.MaxInt, "0 0 list has more than 100 pages"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := readPages(tt.want, func(int) ([]int, PageMeta, error) {
				return []int{1}, PageMeta{TotalPages: tt.totalPages, TotalItems: 1000}, nil
			})
			if got := fmt.Sprint(len(items), " ", total, " ", err); got != tt.got {
				t.Errorf("got %q, want %q", got, tt.got)
			}
		})
	}
}
