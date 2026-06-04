package clients

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func cookieClient(t *testing.T, serverURL string) *APIClient {
	t.Helper()
	client, err := NewAPIClient(
		serverURL, 5*time.Second, "", "",
		[]*http.Cookie{{Name: "session", Value: "abc"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}
	return client
}

func TestListMcpConnections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"connections":[{"id":"c1","name":"github","type":"custom","status":"ok","tool_count":3}]}}`,
		)
	}))
	defer server.Close()

	data, err := cookieClient(
		t,
		server.URL,
	).ListMcpConnections(context.Background(), "org-1", "proj-1")
	if err != nil {
		t.Fatalf("ListMcpConnections() error = %v", err)
	}
	if len(data.Connections) != 1 || data.Connections[0].ID != "c1" {
		t.Fatalf("unexpected connections: %#v", data.Connections)
	}
}

func TestGetMcpConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections/c1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"connection":{"id":"c1","name":"github","type":"custom","status":"ok","tools":[{"name":"search","enabled":true}]}}}`,
		)
	}))
	defer server.Close()

	conn, err := cookieClient(
		t,
		server.URL,
	).GetMcpConnection(context.Background(), "org-1", "proj-1", "c1")
	if err != nil {
		t.Fatalf("GetMcpConnection() error = %v", err)
	}
	if conn.ID != "c1" || len(conn.Tools) != 1 || conn.Tools[0].Name != "search" {
		t.Fatalf("unexpected detail: %#v", conn)
	}
}

func TestListMcpCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-catalog" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"entries":[{"id":"e1","name":"GitHub","category":"dev","type":"platform","auth_methods":["api_key"]}]}}`,
		)
	}))
	defer server.Close()

	data, err := cookieClient(t, server.URL).ListMcpCatalog(context.Background(), "org-1", "proj-1")
	if err != nil {
		t.Fatalf("ListMcpCatalog() error = %v", err)
	}
	if len(data.Entries) != 1 || data.Entries[0].ID != "e1" {
		t.Fatalf("unexpected catalog: %#v", data.Entries)
	}
}

func TestCreateMcpConnection(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"connection":{"id":"c9","name":"db","type":"custom","status":"ok"}}}`,
		)
	}))
	defer server.Close()

	conn, err := cookieClient(t, server.URL).CreateMcpConnection(
		context.Background(), "org-1", "proj-1",
		McpConnectionCreateBody{
			Type:        "custom",
			Name:        "db",
			EndpointURL: "https://x",
			AuthType:    "none",
		},
	)
	if err != nil {
		t.Fatalf("CreateMcpConnection() error = %v", err)
	}
	if conn.ID != "c9" {
		t.Fatalf("unexpected connection: %#v", conn)
	}
	if !strings.Contains(gotBody, `"type":"custom"`) ||
		!strings.Contains(gotBody, `"endpoint_url":"https://x"`) {
		t.Fatalf("unexpected request body: %s", gotBody)
	}
}

func TestCreateMcpConnectionFromCatalog(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"connection":{"id":"c5","name":"gh","type":"platform","status":"ok"}}}`,
		)
	}))
	defer server.Close()

	conn, err := cookieClient(t, server.URL).CreateMcpConnection(
		context.Background(), "org-1", "proj-1",
		McpConnectionCreateBody{
			Type:        "platform",
			CatalogID:   "github",
			EndpointURL: "https://mcp.github.com/",
			Name:        "gh",
			AuthType:    "bearer",
			Credential:  "tok",
		},
	)
	if err != nil {
		t.Fatalf("CreateMcpConnection() error = %v", err)
	}
	if conn.ID != "c5" {
		t.Fatalf("unexpected connection: %#v", conn)
	}
	if !strings.Contains(gotBody, `"type":"platform"`) ||
		!strings.Contains(gotBody, `"catalog_id":"github"`) {
		t.Fatalf("unexpected request body: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"endpoint_url":"https://mcp.github.com/"`) {
		t.Fatalf("catalog body should forward endpoint_url: %s", gotBody)
	}
	if strings.Contains(gotBody, "transport") {
		t.Fatalf("catalog body should omit transport: %s", gotBody)
	}
}

func TestDeleteMcpConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections/c1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"success":true,"data":{"id":"c1","deleted":true}}`)
	}))
	defer server.Close()

	if err := cookieClient(
		t,
		server.URL,
	).DeleteMcpConnection(context.Background(), "org-1", "proj-1", "c1"); err != nil {
		t.Fatalf("DeleteMcpConnection() error = %v", err)
	}
}

func TestVerifyMcpConnection(t *testing.T) {
	var gotBody, gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections/c1/verify" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"status":"ok","protocol_version":"2025-03-26","tools":[{"name":"t","enabled":true}]}}`,
		)
	}))
	defer server.Close()

	res, err := cookieClient(
		t,
		server.URL,
	).VerifyMcpConnection(context.Background(), "org-1", "proj-1", "c1")
	if err != nil {
		t.Fatalf("VerifyMcpConnection() error = %v", err)
	}
	if res.Status != "ok" || len(res.Tools) != 1 {
		t.Fatalf("unexpected verify: %#v", res)
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	if strings.TrimSpace(gotBody) != "{}" {
		t.Fatalf("verify body = %q, want {}", gotBody)
	}
}

func TestRunMcpTool(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections/c1/tools/search/run" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"status":"ok","result":{"content":"hi"}}}`,
		)
	}))
	defer server.Close()

	res, err := cookieClient(t, server.URL).RunMcpTool(
		context.Background(), "org-1", "proj-1", "c1", "search",
		map[string]any{"q": "foo"},
	)
	if err != nil {
		t.Fatalf("RunMcpTool() error = %v", err)
	}
	if res.Status != "ok" {
		t.Fatalf("unexpected tool result: %#v", res)
	}
	if !strings.Contains(gotBody, `"arguments":{"q":"foo"}`) {
		t.Fatalf("unexpected request body: %s", gotBody)
	}
}

func TestRunMcpToolNilArgs(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, `{"success":true,"data":{"status":"ok"}}`)
	}))
	defer server.Close()

	_, err := cookieClient(t, server.URL).RunMcpTool(
		context.Background(), "org-1", "proj-1", "c1", "search", nil,
	)
	if err != nil {
		t.Fatalf("RunMcpTool() error = %v", err)
	}
	if !strings.Contains(gotBody, `"arguments":{}`) {
		t.Fatalf("nil args should send empty object, got: %s", gotBody)
	}
}
