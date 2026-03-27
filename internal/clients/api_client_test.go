package clients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewAPIClient(t *testing.T) {
	t.Run("creates client with cookies", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.isApiKeyMode {
			t.Error("expected isApiKeyMode to be false")
		}
	})

	t.Run("returns error with no auth", func(t *testing.T) {
		_, err := NewAPIClient("https://api.example.com", 30*time.Second, "", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no authentication method available") {
			t.Errorf("error should mention 'no authentication method available', got: %v", err)
		}
	})
}

func TestAPIClientGetOrgIdByName(t *testing.T) {
	t.Run("returns error for empty name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetOrgIdByName(ctx, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "organization name cannot be empty") {
			t.Errorf("error should mention 'organization name cannot be empty', got: %v", err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetOrgIdByName(ctx, "   ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "organization name cannot be empty") {
			t.Errorf("error should mention 'organization name cannot be empty', got: %v", err)
		}
	})
}

func TestAPIClientGetProjectByName(t *testing.T) {
	t.Run("returns error for empty name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetProjectByName(ctx, "org-123", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "project name cannot be empty") {
			t.Errorf("error should mention 'project name cannot be empty', got: %v", err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetProjectByName(ctx, "org-123", "   ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "project name cannot be empty") {
			t.Errorf("error should mention 'project name cannot be empty', got: %v", err)
		}
	})
}

func TestAPIClientGetProjectId(t *testing.T) {
	t.Run("returns error for empty org name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, _, err = client.GetProjectId(ctx, "", "project")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "organization name cannot be empty") {
			t.Errorf("error should mention 'organization name cannot be empty', got: %v", err)
		}
	})

	t.Run("returns error for empty project name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, _, err = client.GetProjectId(ctx, "org", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "project name cannot be empty") {
			t.Errorf("error should mention 'project name cannot be empty', got: %v", err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, _, err = client.GetProjectId(ctx, "  ", "  ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAPIClientSearchObservations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/observations" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("from_start_time") != "2025-01-01T00:00:00Z" {
			t.Fatalf("from_start_time = %q", query.Get("from_start_time"))
		}
		if query.Get("limit") != "20" {
			t.Fatalf("limit = %q", query.Get("limit"))
		}
		if query.Get("trace_id") != "trace-1" {
			t.Fatalf("trace_id = %q", query.Get("trace_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"observations":[{"id":"obs-1","trace_id":"trace-1"}],"meta":{"next_cursor":"cursor-2"}}}`,
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	observations, meta, _, err := client.SearchObservations(
		context.Background(),
		"org-1",
		"proj-1",
		ObservationSearchOptions{
			FromTimestamp: "2025-01-01T00:00:00Z",
			Limit:         20,
			TraceID:       "trace-1",
		},
	)
	if err != nil {
		t.Fatalf("SearchObservations() error = %v", err)
	}
	if len(observations) != 1 || observations[0].ID != "obs-1" {
		t.Fatalf("unexpected observations: %#v", observations)
	}
	if meta.NextCursor != "cursor-2" {
		t.Fatalf("next cursor = %q", meta.NextCursor)
	}
}

func TestAPIClientListSessions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/sessions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("page = %q", r.URL.Query().Get("page"))
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"sessions":[{"id":"sess-1"}],"meta":{"page":2,"limit":10,"total_items":11,"total_pages":2}}}`,
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	sessions, meta, _, err := client.ListSessions(
		context.Background(),
		"org-1",
		"proj-1",
		SessionListOptions{Page: 2, Limit: 10},
	)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "sess-1" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
	if meta.Page != 2 || meta.TotalItems != 11 {
		t.Fatalf("unexpected meta: %#v", meta)
	}
}

func TestAPIClientListScores(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/scores" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("cursor") != "cursor-1" {
			t.Fatalf("cursor = %q", query.Get("cursor"))
		}
		if query.Get("trace_id") != "trace-1" {
			t.Fatalf("trace_id = %q", query.Get("trace_id"))
		}
		if query.Get("trace_tag") != "tag-1" {
			t.Fatalf("trace_tag = %q", query.Get("trace_tag"))
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"scores":[{"id":"score-1","value":0.95}],"meta":{"next_cursor":"cursor-2"}}}`,
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	scores, meta, _, err := client.ListScores(
		context.Background(),
		"org-1",
		"proj-1",
		ScoreListOptions{
			Cursor:    "cursor-1",
			TraceID:   "trace-1",
			TraceTags: []string{"tag-1"},
		},
	)
	if err != nil {
		t.Fatalf("ListScores() error = %v", err)
	}
	if len(scores) != 1 || scores[0].ID != "score-1" {
		t.Fatalf("unexpected scores: %#v", scores)
	}
	if meta.NextCursor != "cursor-2" {
		t.Fatalf("next cursor = %q", meta.NextCursor)
	}
}

func TestAPIClientCreateScoreRequiresAPIKey(t *testing.T) {
	client, err := NewAPIClient(
		"https://api.example.com",
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	_, _, err = client.CreateScore(context.Background(), "org-1", "proj-1", ScoreCreateBody{
		Name:  "quality",
		Value: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--api-key") {
		t.Fatalf("error = %v, want api key guidance", err)
	}
}

func TestAPIClientCreateScoreAndDeleteTraces(t *testing.T) {
	var capturedCreateBody ScoreCreateBody
	var capturedDeleteBody BulkTraceDeleteBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/validate-api-key":
			w.Header().Set("x-org-id", "org-1")
			w.Header().Set("x-org-name", "Org 1")
			w.Header().Set("x-project-id", "proj-1")
			w.Header().Set("x-project-name", "Project 1")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/platform/v1/organizations/org-1/projects/proj-1/scores":
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &capturedCreateBody); err != nil {
				t.Fatalf("failed to decode create body: %v", err)
			}
			_, _ = io.WriteString(
				w,
				`{"success":true,"data":{"score":{"id":"score-1","name":"quality","value":1}}}`,
			)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/platform/v1/organizations/org-1/projects/proj-1/traces":
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &capturedDeleteBody); err != nil {
				t.Fatalf("failed to decode delete body: %v", err)
			}
			_, _ = io.WriteString(w, `{"success":true,"message":"Deleted 2 traces."}`)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, 5*time.Second, "api-key", nil)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	score, _, err := client.CreateScore(context.Background(), "org-1", "proj-1", ScoreCreateBody{
		Name:    "quality",
		TraceID: "trace-1",
		Value:   1,
	})
	if err != nil {
		t.Fatalf("CreateScore() error = %v", err)
	}
	if score.ID != "score-1" {
		t.Fatalf("score ID = %q", score.ID)
	}
	if capturedCreateBody.TraceID != "trace-1" || capturedCreateBody.Name != "quality" {
		t.Fatalf("captured create body = %#v", capturedCreateBody)
	}

	message, err := client.DeleteTraces(
		context.Background(),
		"org-1",
		"proj-1",
		BulkTraceDeleteBody{
			IDs: []string{"trace-1", "trace-2"},
		},
	)
	if err != nil {
		t.Fatalf("DeleteTraces() error = %v", err)
	}
	if message != "Deleted 2 traces." {
		t.Fatalf("message = %q", message)
	}
	if len(capturedDeleteBody.IDs) != 2 {
		t.Fatalf("captured delete body = %#v", capturedDeleteBody)
	}
}

func TestAPIClientGetSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/sessions/sess-1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("fields") != "core,traces" {
			t.Fatalf("fields = %q", r.URL.Query().Get("fields"))
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"session":{"id":"sess-1","created_at":"2025-01-01T00:00:00Z","traces":[{"id":"trace-1","name":"t1"}]}}}`,
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	session, _, err := client.GetSession(
		context.Background(),
		"org-1",
		"proj-1",
		"sess-1",
		"core,traces",
	)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.ID != "sess-1" {
		t.Fatalf("session ID = %q", session.ID)
	}
	if len(session.Traces) != 1 || session.Traces[0].ID != "trace-1" {
		t.Fatalf("unexpected traces: %#v", session.Traces)
	}
}

func TestAPIClientDeleteScore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/validate-api-key":
			w.Header().Set("x-org-id", "org-1")
			w.Header().Set("x-org-name", "Org 1")
			w.Header().Set("x-project-id", "proj-1")
			w.Header().Set("x-project-name", "Project 1")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/platform/v1/organizations/org-1/projects/proj-1/scores/score-1":
			_, _ = io.WriteString(w, `{"success":true,"message":"Score deleted."}`)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, 5*time.Second, "api-key", nil)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	message, err := client.DeleteScore(context.Background(), "org-1", "proj-1", "score-1")
	if err != nil {
		t.Fatalf("DeleteScore() error = %v", err)
	}
	if message != "Score deleted." {
		t.Fatalf("message = %q", message)
	}
}

func TestAPIClientDeleteTrace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/validate-api-key":
			w.Header().Set("x-org-id", "org-1")
			w.Header().Set("x-org-name", "Org 1")
			w.Header().Set("x-project-id", "proj-1")
			w.Header().Set("x-project-name", "Project 1")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/platform/v1/organizations/org-1/projects/proj-1/traces/trace-1":
			_, _ = io.WriteString(w, `{"success":true,"message":"Trace deleted."}`)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, 5*time.Second, "api-key", nil)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	message, err := client.DeleteTrace(context.Background(), "org-1", "proj-1", "trace-1")
	if err != nil {
		t.Fatalf("DeleteTrace() error = %v", err)
	}
	if message != "Trace deleted." {
		t.Fatalf("message = %q", message)
	}
}

func TestAPIClientListMetricsDailySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/metrics/daily" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("from_timestamp") != "2025-01-01T00:00:00Z" {
			t.Fatalf("from_timestamp = %q", query.Get("from_timestamp"))
		}
		if query.Get("trace_name") != "my-trace" {
			t.Fatalf("trace_name = %q", query.Get("trace_name"))
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"metrics":[{"date":"2025-01-01","count_traces":10,"count_observations":50,"total_cost":1.5}],"meta":{"page":1,"limit":50,"total_items":1,"total_pages":1}}}`,
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	metrics, meta, _, err := client.ListMetricsDaily(
		context.Background(),
		"org-1",
		"proj-1",
		MetricsDailyOptions{
			FromTimestamp: "2025-01-01T00:00:00Z",
			TraceName:     "my-trace",
		},
	)
	if err != nil {
		t.Fatalf("ListMetricsDaily() error = %v", err)
	}
	if len(metrics) != 1 || metrics[0].Date != "2025-01-01" {
		t.Fatalf("unexpected metrics: %#v", metrics)
	}
	if meta.Page != 1 || meta.TotalItems != 1 {
		t.Fatalf("unexpected meta: %#v", meta)
	}
}

func TestAPIClientListMetricsDailyReturnsServerMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"success":false,"error":{"message":"bad filter"}}`)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}

	_, _, _, err = client.ListMetricsDaily(
		context.Background(),
		"org-1",
		"proj-1",
		MetricsDailyOptions{},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "bad filter") {
		t.Fatalf("error = %v, want extracted server message", err)
	}
}
