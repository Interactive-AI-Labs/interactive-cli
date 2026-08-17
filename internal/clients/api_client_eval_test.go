package clients

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
	wantQuery map[string]string,
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
		query := r.URL.Query()
		for key, want := range wantQuery {
			if got := query.Get(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		_, _ = io.WriteString(w, body)
	}))
}

func TestAPIClientListAnnotationQueues(t *testing.T) {
	const path = "/api/platform/v1/organizations/org-1/projects/proj-1/annotation-queues"
	const body = `{"success":true,"data":{"queues":[{"id":"q-1","name":"queue",` +
		`"count_completed_items":2,"count_pending_items":3}],` +
		`"meta":{"page":1,"total_pages":1,"total_items":1}}}`

	tests := []struct {
		name      string
		opts      AnnotationQueueListOptions
		wantQuery map[string]string
	}{
		{
			name:      "omits sorting when unset",
			opts:      AnnotationQueueListOptions{Page: 1},
			wantQuery: map[string]string{"sort_by": "", "sort_order": ""},
		},
		{
			name:      "sends sort_by alone",
			opts:      AnnotationQueueListOptions{Page: 1, SortBy: "name"},
			wantQuery: map[string]string{"sort_by": "name", "sort_order": ""},
		},
		{
			name:      "sends sort_by and sort_order",
			opts:      AnnotationQueueListOptions{Page: 1, SortBy: "name", SortOrder: "asc"},
			wantQuery: map[string]string{"sort_by": "name", "sort_order": "asc"},
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
		wantQuery map[string]string
	}{
		{
			name:      "omits sorting when unset",
			opts:      QueueItemListOptions{Page: 1},
			wantQuery: map[string]string{"status": "", "sort_by": "", "sort_order": ""},
		},
		{
			name:      "sends sort_by alone",
			opts:      QueueItemListOptions{Page: 1, SortBy: "status"},
			wantQuery: map[string]string{"sort_by": "status", "sort_order": ""},
		},
		{
			name: "sends status alongside sorting",
			opts: QueueItemListOptions{
				Page:      1,
				Status:    "PENDING",
				SortBy:    "completed_at",
				SortOrder: "desc",
			},
			wantQuery: map[string]string{
				"status":     "PENDING",
				"sort_by":    "completed_at",
				"sort_order": "desc",
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
