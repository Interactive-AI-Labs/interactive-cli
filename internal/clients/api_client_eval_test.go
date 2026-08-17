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

func TestAPIClientListAnnotationQueues(t *testing.T) {
	tests := []struct {
		name          string
		opts          AnnotationQueueListOptions
		wantSortBy    string
		wantSortOrder string
	}{
		{
			name: "omits sorting when unset",
			opts: AnnotationQueueListOptions{Page: 1},
		},
		{
			name:          "sends sort_by and sort_order",
			opts:          AnnotationQueueListOptions{Page: 1, SortBy: "name", SortOrder: "asc"},
			wantSortBy:    "name",
			wantSortOrder: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						t.Fatalf("method = %s, want GET", r.Method)
					}
					wantPath := "/api/platform/v1/organizations/org-1/projects/proj-1/annotation-queues"
					if r.URL.Path != wantPath {
						t.Fatalf("path = %s, want %s", r.URL.Path, wantPath)
					}
					query := r.URL.Query()
					if got := query.Get("sort_by"); got != tt.wantSortBy {
						t.Fatalf("sort_by = %q, want %q", got, tt.wantSortBy)
					}
					if got := query.Get("sort_order"); got != tt.wantSortOrder {
						t.Fatalf("sort_order = %q, want %q", got, tt.wantSortOrder)
					}
					_, _ = io.WriteString(
						w,
						`{"success":true,"data":{"queues":[{"id":"q-1","name":"queue"}],`+
							`"meta":{"page":1,"total_pages":1,"total_items":1}}}`,
					)
				}),
			)
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
			if meta.TotalItems != 1 {
				t.Fatalf("unexpected meta: %#v", meta)
			}
		})
	}
}

func TestAPIClientListQueueItems(t *testing.T) {
	tests := []struct {
		name          string
		opts          QueueItemListOptions
		wantStatus    string
		wantSortBy    string
		wantSortOrder string
	}{
		{
			name: "omits sorting when unset",
			opts: QueueItemListOptions{Page: 1},
		},
		{
			name: "sends status alongside sorting",
			opts: QueueItemListOptions{
				Page:      1,
				Status:    "PENDING",
				SortBy:    "completed_at",
				SortOrder: "desc",
			},
			wantStatus:    "PENDING",
			wantSortBy:    "completed_at",
			wantSortOrder: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					wantPath := "/api/platform/v1/organizations/org-1/projects/proj-1/" +
						"annotation-queues/queue-1/items"
					if r.URL.Path != wantPath {
						t.Fatalf("path = %s, want %s", r.URL.Path, wantPath)
					}
					query := r.URL.Query()
					if got := query.Get("status"); got != tt.wantStatus {
						t.Fatalf("status = %q, want %q", got, tt.wantStatus)
					}
					if got := query.Get("sort_by"); got != tt.wantSortBy {
						t.Fatalf("sort_by = %q, want %q", got, tt.wantSortBy)
					}
					if got := query.Get("sort_order"); got != tt.wantSortOrder {
						t.Fatalf("sort_order = %q, want %q", got, tt.wantSortOrder)
					}
					_, _ = io.WriteString(
						w,
						`{"success":true,"data":{"items":[{"id":"i-1","created_at":"2025-01-01"}],`+
							`"meta":{"page":1,"total_pages":1,"total_items":1}}}`,
					)
				}),
			)
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
