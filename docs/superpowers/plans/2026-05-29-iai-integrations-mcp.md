# `iai integrations` MCP Connections Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an `iai integrations` command group that drives the platform's MCP integration-connection API (list, get, create-custom, create-from-catalog, delete, verify, catalog, tools run).

**Architecture:** New Cobra command group in `cmd/integrations.go` reusing `resolveProject` and `groupInfra`; new `APIClient` methods built on the existing generic helpers (`doGet`/`doCreate`/`doList`/`doDelete`/`decodeSuccess`); new table/detail printers in `internal/output/integrations.go` reusing `PrintTable`/`NewDescribeWriter`. CLI docs regenerated via `make docs`.

**Tech Stack:** Go 1.25, Cobra, `net/http`, `net/http/httptest` for tests, `tabwriter`-based output helpers.

**Auth note (load-bearing):** MCP client methods must **NOT** call `requireAPIKeyMode()` (unlike `CreateScore`). The backend MCP endpoints accept session-cookie and bearer-token auth with RBAC (`integrations:read`/`integrations:write`), and the copilot authenticates with a **bearer JWT**. Gating on API-key mode would break both `iai login` (cookie) users and the copilot. Let `ApplyAuth` use whatever credential is configured.

**Backend base path:** `/api/platform/v1/organizations/{org}/projects/{project}` — both IDs come from `resolveProject` (`pCtx.orgId`, `pCtx.projectId`). Responses use the `{success, data}` envelope.

---

## File Structure

| File | Responsibility |
|---|---|
| `internal/clients/mcp.go` | New — MCP Go types, `mcpBasePath`, and the 7 `APIClient` methods. (Kept in its own file rather than growing `api_client.go`, which is already ~1400 lines.) |
| `internal/clients/mcp_test.go` | New — `httptest` tests for the 7 methods + envelope decoding. |
| `internal/output/integrations.go` | New — 5 printers. |
| `internal/output/integrations_test.go` | New — printer golden tests. |
| `cmd/integrations.go` | New — command group + 8 commands + validation helpers. |
| `cmd/integrations_test.go` | New — validation-helper tests. |
| `docs/iai_integrations*.md` | Regenerated via `make docs`. |

The methods live on the existing `*APIClient` type (defined in `internal/clients/api_client.go`); putting them in a new file in the same `clients` package is allowed in Go and keeps the diff focused.

---

## Task 1: MCP Go types and base-path helper

**Files:**
- Create: `internal/clients/mcp.go`

- [ ] **Step 1: Create the types file**

Create `internal/clients/mcp.go`:

```go
package clients

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// --- Read shapes (mirror backend api/platform/v1/schemas/mcp_connections.py).
// The credential is never returned by the API, so it has no field here.

type ConnectedAgentRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type McpTool struct {
	Name        string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
	Enabled     bool           `json:"enabled"`
}

type McpConnection struct {
	ID              string              `json:"id"`
	ProjectID       string              `json:"project_id"`
	CatalogID       string              `json:"catalog_id,omitempty"`
	Type            string              `json:"type"`
	Name            string              `json:"name"`
	Slug            string              `json:"slug"`
	Description     string              `json:"description,omitempty"`
	EndpointURL     string              `json:"endpoint_url"`
	Transport       string              `json:"transport"`
	AuthType        string              `json:"auth_type"`
	HasCredential   bool                `json:"has_credential"`
	CustomHeaders   map[string]string   `json:"custom_headers"`
	Status          string              `json:"status"`
	LastVerifiedAt  string              `json:"last_verified_at,omitempty"`
	LastErrorClass  string              `json:"last_error_class,omitempty"`
	ProtocolVersion string              `json:"protocol_version,omitempty"`
	ToolCount       int                 `json:"tool_count"`
	ConnectedAgents []ConnectedAgentRef `json:"connected_agents"`
	CreatedBy       string              `json:"created_by"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

type McpConnectionDetail struct {
	McpConnection
	Tools     []McpTool `json:"tools"`
	LastError string    `json:"last_error,omitempty"`
}

type McpConnectionListData struct {
	Connections []McpConnection `json:"connections"`
}

type McpConnectionDetailData struct {
	Connection McpConnectionDetail `json:"connection"`
}

type McpConnectionDeleteData struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

type McpVerifyData struct {
	Status          string         `json:"status"`
	ErrorClass      string         `json:"error_class,omitempty"`
	ErrorMessage    string         `json:"error_message,omitempty"`
	ProtocolVersion string         `json:"protocol_version,omitempty"`
	ServerInfo      map[string]any `json:"server_info,omitempty"`
	Tools           []McpTool      `json:"tools"`
}

type McpToolCallData struct {
	Status       string         `json:"status"`
	Result       map[string]any `json:"result,omitempty"`
	ErrorClass   string         `json:"error_class,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

type McpCatalogEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	IconKey     string   `json:"icon_key,omitempty"`
	EndpointURL string   `json:"endpoint_url,omitempty"`
	DocsURL     string   `json:"docs_url,omitempty"`
	AuthMethods []string `json:"auth_methods"`
}

type McpCatalogListData struct {
	Entries []McpCatalogEntry `json:"entries"`
}

// --- Create request. One body covers both connection types; the two
// create-* commands populate it differently.

type McpConnectionCreateBody struct {
	Type          string            `json:"type"`
	CatalogID     string            `json:"catalog_id,omitempty"`
	Name          string            `json:"name"`
	Slug          string            `json:"slug,omitempty"`
	Description   string            `json:"description,omitempty"`
	EndpointURL   string            `json:"endpoint_url,omitempty"`
	Transport     string            `json:"transport,omitempty"`
	AuthType      string            `json:"auth_type"`
	Credential    string            `json:"credential,omitempty"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// mcpBasePath builds the org+project-scoped platform prefix, mirroring
// promptBasePath in api_client.go.
func mcpBasePath(orgID, projectID string) string {
	return fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
	)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go build ./internal/clients/`
Expected: builds clean (imports `context`, `net/http`, `net/url` are used by methods added in Task 2/3; if the build complains about unused imports at this step, that is expected — proceed to Task 2 which adds the consumers, then build).

> Note: to avoid an unused-import failure on this intermediate step, you may add the methods in Task 2 before first building. If building Task 1 alone, temporarily omit the `context`/`net/http`/`net/url` imports and add them with their first use.

- [ ] **Step 3: Commit**

```bash
git add internal/clients/mcp.go
git commit -m "feat(integrations): add mcp client types and base-path helper"
```

---

## Task 2: Read client methods (List, Get, Catalog)

**Files:**
- Modify: `internal/clients/mcp.go`
- Create: `internal/clients/mcp_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/clients/mcp_test.go`:

```go
package clients

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
		_, _ = io.WriteString(w, `{"success":true,"data":{"connections":[{"id":"c1","name":"github","type":"custom","status":"ok","tool_count":3}]}}`)
	}))
	defer server.Close()

	data, err := cookieClient(t, server.URL).ListMcpConnections(context.Background(), "org-1", "proj-1")
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
		_, _ = io.WriteString(w, `{"success":true,"data":{"connection":{"id":"c1","name":"github","type":"custom","status":"ok","tools":[{"name":"search","enabled":true}]}}}`)
	}))
	defer server.Close()

	conn, err := cookieClient(t, server.URL).GetMcpConnection(context.Background(), "org-1", "proj-1", "c1")
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
		_, _ = io.WriteString(w, `{"success":true,"data":{"entries":[{"id":"e1","name":"GitHub","category":"dev","type":"platform","auth_methods":["api_key"]}]}}`)
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./internal/clients/ -run 'TestListMcpConnections|TestGetMcpConnection|TestListMcpCatalog' -v`
Expected: compile error / FAIL — `ListMcpConnections` etc. undefined.

- [ ] **Step 3: Implement the read methods**

Append to `internal/clients/mcp.go`:

```go
func (c *APIClient) ListMcpConnections(
	ctx context.Context, orgID, projectID string,
) (*McpConnectionListData, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections"
	data, _, err := doGet[McpConnectionListData](c, ctx, path, "list mcp connections")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *APIClient) GetMcpConnection(
	ctx context.Context, orgID, projectID, id string,
) (*McpConnectionDetail, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" + url.PathEscape(id)
	data, _, err := doGet[McpConnectionDetailData](c, ctx, path, "get mcp connection")
	if err != nil {
		return nil, err
	}
	return &data.Connection, nil
}

func (c *APIClient) ListMcpCatalog(
	ctx context.Context, orgID, projectID string,
) (*McpCatalogListData, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-catalog"
	data, _, err := doGet[McpCatalogListData](c, ctx, path, "list mcp catalog")
	if err != nil {
		return nil, err
	}
	return &data, nil
}
```

(Add `_ = http.MethodGet` is not needed; `net/http` is used by Task 3 methods. If building only Task 2, the `net/http` import is still unused — Task 3 consumes it. To keep intermediate builds green, implement Task 3 methods in the same editing pass, or drop `net/http` from imports until Task 3.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./internal/clients/ -run 'TestListMcpConnections|TestGetMcpConnection|TestListMcpCatalog' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/clients/mcp.go internal/clients/mcp_test.go
git commit -m "feat(integrations): add mcp read client methods"
```

---

## Task 3: Write client methods (Create, Delete, Verify, RunTool)

**Files:**
- Modify: `internal/clients/mcp.go`
- Modify: `internal/clients/mcp_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/clients/mcp_test.go`:

```go
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
		_, _ = io.WriteString(w, `{"success":true,"data":{"connection":{"id":"c9","name":"db","type":"custom","status":"ok"}}}`)
	}))
	defer server.Close()

	conn, err := cookieClient(t, server.URL).CreateMcpConnection(
		context.Background(), "org-1", "proj-1",
		McpConnectionCreateBody{Type: "custom", Name: "db", EndpointURL: "https://x", AuthType: "none"},
	)
	if err != nil {
		t.Fatalf("CreateMcpConnection() error = %v", err)
	}
	if conn.ID != "c9" {
		t.Fatalf("unexpected connection: %#v", conn)
	}
	if !strings.Contains(gotBody, `"type":"custom"`) || !strings.Contains(gotBody, `"endpoint_url":"https://x"`) {
		t.Fatalf("unexpected request body: %s", gotBody)
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

	if err := cookieClient(t, server.URL).DeleteMcpConnection(context.Background(), "org-1", "proj-1", "c1"); err != nil {
		t.Fatalf("DeleteMcpConnection() error = %v", err)
	}
}

func TestVerifyMcpConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/mcp-connections/c1/verify" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"success":true,"data":{"status":"ok","protocol_version":"2025-03-26","tools":[{"name":"t","enabled":true}]}}`)
	}))
	defer server.Close()

	res, err := cookieClient(t, server.URL).VerifyMcpConnection(context.Background(), "org-1", "proj-1", "c1")
	if err != nil {
		t.Fatalf("VerifyMcpConnection() error = %v", err)
	}
	if res.Status != "ok" || len(res.Tools) != 1 {
		t.Fatalf("unexpected verify: %#v", res)
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
		_, _ = io.WriteString(w, `{"success":true,"data":{"status":"ok","result":{"content":"hi"}}}`)
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
```

Add `"strings"` to the test file's import block.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./internal/clients/ -run 'TestCreateMcpConnection|TestDeleteMcpConnection|TestVerifyMcpConnection|TestRunMcpTool' -v`
Expected: compile error / FAIL — methods undefined.

- [ ] **Step 3: Implement the write methods**

Append to `internal/clients/mcp.go`. These deliberately do **not** call `requireAPIKeyMode()` (see Auth note in the header):

```go
func (c *APIClient) CreateMcpConnection(
	ctx context.Context, orgID, projectID string, body McpConnectionCreateBody,
) (*McpConnectionDetail, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections"
	data, _, err := doCreate[McpConnectionDetailData](c, ctx, path, body, "create mcp connection")
	if err != nil {
		return nil, err
	}
	return &data.Connection, nil
}

func (c *APIClient) DeleteMcpConnection(
	ctx context.Context, orgID, projectID, id string,
) error {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" + url.PathEscape(id)
	_, err := c.doDelete(ctx, path, "delete mcp connection")
	return err
}

func (c *APIClient) VerifyMcpConnection(
	ctx context.Context, orgID, projectID, id string,
) (*McpVerifyData, error) {
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" + url.PathEscape(id) + "/verify"
	// POST with no body; doCreate handles a nil body via newJSONRequest.
	data, _, err := doCreate[McpVerifyData](c, ctx, path, nil, "verify mcp connection")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *APIClient) RunMcpTool(
	ctx context.Context, orgID, projectID, id, tool string, arguments map[string]any,
) (*McpToolCallData, error) {
	if arguments == nil {
		arguments = map[string]any{}
	}
	path := mcpBasePath(orgID, projectID) + "/mcp-connections/" +
		url.PathEscape(id) + "/tools/" + url.PathEscape(tool) + "/run"
	body := struct {
		Arguments map[string]any `json:"arguments"`
	}{Arguments: arguments}
	data, _, err := doCreate[McpToolCallData](c, ctx, path, body, "run mcp tool")
	if err != nil {
		return nil, err
	}
	return &data, nil
}
```

Now confirm `net/http` is actually referenced in `mcp.go`; it is not (the helpers wrap it). Remove `net/http` from the `mcp.go` import block — keep only `context`, `fmt`, `net/url`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./internal/clients/ -v`
Expected: PASS — all MCP tests plus the pre-existing client tests.

- [ ] **Step 5: Commit**

```bash
git add internal/clients/mcp.go internal/clients/mcp_test.go
git commit -m "feat(integrations): add mcp write client methods"
```

---

## Task 4: Output printers

**Files:**
- Create: `internal/output/integrations.go`
- Create: `internal/output/integrations_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/output/integrations_test.go`:

```go
package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintMcpConnectionListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintMcpConnectionList(&buf, nil); err != nil {
		t.Fatalf("error = %v", err)
	}
	if buf.String() != "No integration connections found.\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPrintMcpConnectionListRows(t *testing.T) {
	var buf bytes.Buffer
	conns := []clients.McpConnection{
		{Name: "github", Type: "custom", Status: "ok", ToolCount: 3, EndpointURL: "https://api.githubcopilot.com/mcp", UpdatedAt: "2026-05-01T10:00:00Z"},
	}
	if err := PrintMcpConnectionList(&buf, conns); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "github") || !strings.Contains(out, "custom") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestPrintMcpConnectionDetailWithTools(t *testing.T) {
	var buf bytes.Buffer
	conn := &clients.McpConnectionDetail{
		McpConnection: clients.McpConnection{
			ID: "c1", Name: "github", Type: "custom", Status: "ok",
			EndpointURL: "https://x", Transport: "streamable_http", AuthType: "bearer",
		},
		Tools: []clients.McpTool{{Name: "search", Enabled: true, Description: "Search repos"}},
	}
	if err := PrintMcpConnectionDetail(&buf, conn); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "github") || !strings.Contains(out, "search") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestPrintMcpCatalog(t *testing.T) {
	var buf bytes.Buffer
	entries := []clients.McpCatalogEntry{
		{ID: "e1", Name: "GitHub", Category: "dev", Type: "platform", AuthMethods: []string{"api_key"}},
	}
	if err := PrintMcpCatalog(&buf, entries); err != nil {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(buf.String(), "GitHub") || !strings.Contains(buf.String(), "e1") {
		t.Fatalf("unexpected output:\n%s", buf.String())
	}
}

func TestPrintMcpVerifyResultError(t *testing.T) {
	var buf bytes.Buffer
	res := &clients.McpVerifyData{Status: "error", ErrorClass: "unauthorized", ErrorMessage: "bad token"}
	if err := PrintMcpVerifyResult(&buf, res); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "error") || !strings.Contains(out, "unauthorized") || !strings.Contains(out, "bad token") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestPrintMcpToolResultOk(t *testing.T) {
	var buf bytes.Buffer
	res := &clients.McpToolCallData{Status: "ok", Result: map[string]any{"content": "hi"}}
	if err := PrintMcpToolResult(&buf, res); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ok") || !strings.Contains(out, "content") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./internal/output/ -run 'Mcp' -v`
Expected: compile error / FAIL — printers undefined.

- [ ] **Step 3: Implement the printers**

Create `internal/output/integrations.go`:

```go
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintMcpConnectionList(out io.Writer, conns []clients.McpConnection) error {
	if len(conns) == 0 {
		fmt.Fprintln(out, "No integration connections found.")
		return nil
	}
	headers := []string{"ID", "NAME", "TYPE", "STATUS", "TOOLS", "ENDPOINT", "UPDATED"}
	rows := make([][]string, len(conns))
	for i, c := range conns {
		rows[i] = []string{
			c.ID,
			c.Name,
			c.Type,
			c.Status,
			fmt.Sprintf("%d", c.ToolCount),
			c.EndpointURL,
			LocalTime(c.UpdatedAt),
		}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpConnectionDetail(out io.Writer, conn *clients.McpConnectionDetail) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", conn.ID)
	fmt.Fprintf(w, "Name:\t%s\n", conn.Name)
	fmt.Fprintf(w, "Type:\t%s\n", conn.Type)
	fmt.Fprintf(w, "Status:\t%s\n", conn.Status)
	if conn.Slug != "" {
		fmt.Fprintf(w, "Slug:\t%s\n", conn.Slug)
	}
	if conn.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", conn.Description)
	}
	fmt.Fprintf(w, "Endpoint:\t%s\n", conn.EndpointURL)
	fmt.Fprintf(w, "Transport:\t%s\n", conn.Transport)
	fmt.Fprintf(w, "Auth Type:\t%s\n", conn.AuthType)
	if conn.CatalogID != "" {
		fmt.Fprintf(w, "Catalog ID:\t%s\n", conn.CatalogID)
	}
	if conn.ProtocolVersion != "" {
		fmt.Fprintf(w, "Protocol:\t%s\n", conn.ProtocolVersion)
	}
	if conn.LastVerifiedAt != "" {
		fmt.Fprintf(w, "Last Verified:\t%s\n", LocalTime(conn.LastVerifiedAt))
	}
	if conn.LastError != "" {
		fmt.Fprintf(w, "Last Error:\t%s (%s)\n", conn.LastError, conn.LastErrorClass)
	}
	if len(conn.ConnectedAgents) > 0 {
		names := make([]string, len(conn.ConnectedAgents))
		for i, a := range conn.ConnectedAgents {
			names[i] = a.Name
		}
		fmt.Fprintf(w, "Connected Agents:\t%s\n", strings.Join(names, ", "))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if len(conn.Tools) == 0 {
		fmt.Fprintln(out, "\nNo tools discovered yet. Run 'iai integrations verify' to refresh.")
		return nil
	}
	fmt.Fprintln(out, "\nTools:")
	headers := []string{"NAME", "ENABLED", "DESCRIPTION"}
	rows := make([][]string, len(conn.Tools))
	for i, tl := range conn.Tools {
		rows[i] = []string{tl.Name, fmt.Sprintf("%t", tl.Enabled), tl.Description}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpCatalog(out io.Writer, entries []clients.McpCatalogEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(out, "No catalog entries found.")
		return nil
	}
	headers := []string{"ID", "NAME", "CATEGORY", "TYPE", "AUTH"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{e.ID, e.Name, e.Category, e.Type, TruncateList(e.AuthMethods, 3)}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpVerifyResult(out io.Writer, res *clients.McpVerifyData) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Status:\t%s\n", res.Status)
	if res.ProtocolVersion != "" {
		fmt.Fprintf(w, "Protocol:\t%s\n", res.ProtocolVersion)
	}
	if res.ErrorClass != "" {
		fmt.Fprintf(w, "Error Class:\t%s\n", res.ErrorClass)
	}
	if res.ErrorMessage != "" {
		fmt.Fprintf(w, "Error:\t%s\n", res.ErrorMessage)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if len(res.Tools) == 0 {
		return nil
	}
	fmt.Fprintln(out, "\nTools:")
	headers := []string{"NAME", "ENABLED", "DESCRIPTION"}
	rows := make([][]string, len(res.Tools))
	for i, tl := range res.Tools {
		rows[i] = []string{tl.Name, fmt.Sprintf("%t", tl.Enabled), tl.Description}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpToolResult(out io.Writer, res *clients.McpToolCallData) error {
	fmt.Fprintf(out, "Status: %s\n", res.Status)
	if res.ErrorClass != "" {
		fmt.Fprintf(out, "Error Class: %s\n", res.ErrorClass)
	}
	if res.ErrorMessage != "" {
		fmt.Fprintf(out, "Error: %s\n", res.ErrorMessage)
	}
	if res.Result != nil {
		pretty, err := json.MarshalIndent(res.Result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "\nResult:\n%s\n", string(pretty))
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./internal/output/ -run 'Mcp' -v`
Expected: PASS (6 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/output/integrations.go internal/output/integrations_test.go
git commit -m "feat(integrations): add mcp output printers"
```

---

## Task 5: Command group skeleton + `list`, `get`, `catalog`

**Files:**
- Create: `cmd/integrations.go`

- [ ] **Step 1: Create the command group with read commands**

Create `cmd/integrations.go`:

```go
package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	parentCmd := &cobra.Command{
		Use:     "integrations",
		Aliases: []string{"integration", "mcp"},
		Short:   "MCP integration connections for a project",
		GroupID: groupInfra,
		Long: `Manage Model Context Protocol (MCP) integration connections in an InteractiveAI project.

Connections let agents reach external tools exposed by an MCP server — either a
curated catalog entry (a vendor-hosted server) or a custom endpoint you define.
Create a connection, verify it to discover its tools, then run a tool directly.`,
	}

	parentCmd.AddCommand(
		makeIntegrationsListCmd(),
		makeIntegrationsGetCmd(),
		makeIntegrationsCatalogCmd(),
		makeIntegrationsCreateCustomCmd(),
		makeIntegrationsCreateFromCatalogCmd(),
		makeIntegrationsDeleteCmd(),
		makeIntegrationsVerifyCmd(),
		makeIntegrationsToolsCmd(),
	)

	rootCmd.AddCommand(parentCmd)
}

func makeIntegrationsListCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List integration connections in a project",
		Long: `List the MCP integration connections in a project, showing each connection's
type, status, tool count, and endpoint.

Examples:
  iai integrations list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			data, err := apiClient.ListMcpConnections(cmd.Context(), pCtx.orgId, pCtx.projectId)
			if err != nil {
				return err
			}
			return output.PrintMcpConnectionList(out, data.Connections)
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connections")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func makeIntegrationsGetCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:     "get <connection-id>",
		Aliases: []string{"describe", "desc"},
		Short:   "Show an integration connection and its tools",
		Long: `Show detailed information about a single integration connection, including the
cached list of tools discovered from the MCP server.

Examples:
  iai integrations get 3f9c1a2e-...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			conn, err := apiClient.GetMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id)
			if err != nil {
				return err
			}
			return output.PrintMcpConnectionDetail(out, conn)
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func makeIntegrationsCatalogCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Browse the MCP integrations catalog",
		Long: `List the curated catalog of MCP servers you can connect to with
'iai integrations create-from-catalog', showing each entry's id, category, and
supported auth methods.

Examples:
  iai integrations catalog`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			data, err := apiClient.ListMcpCatalog(cmd.Context(), pCtx.orgId, pCtx.projectId)
			if err != nil {
				return err
			}
			return output.PrintMcpCatalog(out, data.Entries)
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name to browse the catalog for")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}
```

> The four `make*` functions referenced but not yet defined (`makeIntegrationsCreateCustomCmd`, `makeIntegrationsCreateFromCatalogCmd`, `makeIntegrationsDeleteCmd`, `makeIntegrationsVerifyCmd`, `makeIntegrationsToolsCmd`) are added in Tasks 6–8. To keep the build green between tasks, add temporary stubs returning `&cobra.Command{Use: "..."}` now, or implement Tasks 6–8 in the same pass before building. The plan assumes same-pass implementation; if building incrementally, stub them.

- [ ] **Step 2: Add temporary stubs so it compiles (only if building before Tasks 6-8)**

Append minimal stubs (delete them as each real function lands):

```go
func makeIntegrationsCreateCustomCmd() *cobra.Command      { return &cobra.Command{Use: "create-custom", RunE: func(*cobra.Command, []string) error { return nil }} }
func makeIntegrationsCreateFromCatalogCmd() *cobra.Command { return &cobra.Command{Use: "create-from-catalog", RunE: func(*cobra.Command, []string) error { return nil }} }
func makeIntegrationsDeleteCmd() *cobra.Command            { return &cobra.Command{Use: "delete", RunE: func(*cobra.Command, []string) error { return nil }} }
func makeIntegrationsVerifyCmd() *cobra.Command            { return &cobra.Command{Use: "verify", RunE: func(*cobra.Command, []string) error { return nil }} }
func makeIntegrationsToolsCmd() *cobra.Command             { return &cobra.Command{Use: "tools", RunE: func(*cobra.Command, []string) error { return nil }} }
```

- [ ] **Step 3: Build and smoke-test help**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go build ./... && go run main.go integrations --help`
Expected: builds; help lists `list`, `get`, `catalog` (+ stubs) under the integrations command.

- [ ] **Step 4: Commit**

```bash
git add cmd/integrations.go
git commit -m "feat(integrations): add integrations command group with list/get/catalog"
```

---

## Task 6: `create-custom` and `create-from-catalog` + validation helpers

**Files:**
- Modify: `cmd/integrations.go`
- Create: `cmd/integrations_test.go`

- [ ] **Step 1: Write the failing validation-helper tests**

Create `cmd/integrations_test.go`:

```go
package cmd

import "testing"

func TestValidateMcpAuth(t *testing.T) {
	tests := []struct {
		name       string
		authType   string
		credential string
		wantErr    bool
	}{
		{"none without credential ok", "none", "", false},
		{"none with credential rejected", "none", "secret", true},
		{"api_key requires credential", "api_key", "", true},
		{"api_key with credential ok", "api_key", "secret", false},
		{"bearer requires credential", "bearer", "", true},
		{"invalid auth type", "oauth", "secret", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMcpAuth(tt.authType, tt.credential)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMcpAuth(%q,%q) err=%v wantErr=%v", tt.authType, tt.credential, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMcpTransport(t *testing.T) {
	if err := validateMcpTransport("streamable_http"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateMcpTransport("sse"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateMcpTransport("grpc"); err == nil {
		t.Fatal("expected error for invalid transport")
	}
}

func TestParseHeaderFlags(t *testing.T) {
	got, err := parseHeaderFlags([]string{"X-A=1", "X-B=two=2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["X-A"] != "1" || got["X-B"] != "two=2" {
		t.Fatalf("unexpected headers: %#v", got)
	}
	if _, err := parseHeaderFlags([]string{"bad-no-equals"}); err == nil {
		t.Fatal("expected error for header without '='")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./cmd/ -run 'TestValidateMcp|TestParseHeaderFlags' -v`
Expected: compile error / FAIL — helpers undefined.

- [ ] **Step 3: Implement validation helpers + the two create commands**

Replace the `create-custom` / `create-from-catalog` stubs in `cmd/integrations.go` with the real implementations, and add the helpers. Update the import block to add `"fmt"` and `"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"`:

```go
var validMcpAuthTypes = []string{"api_key", "bearer", "none"}
var validMcpTransports = []string{"streamable_http", "sse"}

func validateMcpAuth(authType, credential string) error {
	valid := false
	for _, a := range validMcpAuthTypes {
		if authType == a {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid --auth-type %q: must be one of %s", authType, strings.Join(validMcpAuthTypes, ", "))
	}
	if authType == "none" && credential != "" {
		return fmt.Errorf("--credential must not be set when --auth-type is 'none'")
	}
	if authType != "none" && credential == "" {
		return fmt.Errorf("--credential is required when --auth-type is %q", authType)
	}
	return nil
}

func validateMcpTransport(transport string) error {
	for _, tr := range validMcpTransports {
		if transport == tr {
			return nil
		}
	}
	return fmt.Errorf("invalid --transport %q: must be one of %s", transport, strings.Join(validMcpTransports, ", "))
}

func parseHeaderFlags(pairs []string) (map[string]string, error) {
	headers := make(map[string]string, len(pairs))
	for _, p := range pairs {
		key, value, found := strings.Cut(p, "=")
		if !found || key == "" {
			return nil, fmt.Errorf("invalid --header %q: expected KEY=VALUE", p)
		}
		headers[key] = value
	}
	return headers, nil
}

func makeIntegrationsCreateCustomCmd() *cobra.Command {
	var (
		endpointURL string
		authType    string
		credential  string
		transport   string
		slug        string
		description string
		headers     []string
		project     string
		org         string
	)
	cmd := &cobra.Command{
		Use:   "create-custom <name>",
		Short: "Connect a custom MCP endpoint",
		Long: `Create an integration connection to a custom (user-defined) MCP endpoint.

The connection is verified against the live server on save; if the server cannot
be reached or rejects the credential, creation fails and nothing is stored.

Examples:
  iai integrations create-custom my-server \
    --endpoint-url https://mcp.example.com/mcp --auth-type none
  iai integrations create-custom github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential "$GITHUB_TOKEN"
  iai integrations create-custom internal \
    --endpoint-url https://mcp.internal/sse --transport sse \
    --auth-type api_key --credential "$KEY" --header "X-Team=platform"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			if err := validateMcpAuth(authType, credential); err != nil {
				return err
			}
			if err := validateMcpTransport(transport); err != nil {
				return err
			}
			customHeaders, err := parseHeaderFlags(headers)
			if err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			body := clients.McpConnectionCreateBody{
				Type:          "custom",
				Name:          name,
				Slug:          slug,
				Description:   description,
				EndpointURL:   endpointURL,
				Transport:     transport,
				AuthType:      authType,
				Credential:    credential,
				CustomHeaders: customHeaders,
			}

			fmt.Fprintf(out, "\nConnecting %q and verifying...\n\n", name)
			conn, err := apiClient.CreateMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, body)
			if err != nil {
				return err
			}
			return output.PrintMcpConnectionDetail(out, conn)
		},
	}
	cmd.Flags().StringVar(&endpointURL, "endpoint-url", "", "MCP server endpoint URL (required)")
	cmd.Flags().StringVar(&authType, "auth-type", "", "Auth type: api_key, bearer, or none (required)")
	cmd.Flags().StringVar(&credential, "credential", "", "API key or bearer token (required unless auth-type=none)")
	cmd.Flags().StringVar(&transport, "transport", "streamable_http", "Transport: streamable_http (default) or sse")
	cmd.Flags().StringVar(&slug, "slug", "", "Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)")
	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringArrayVar(&headers, "header", nil, "Extra header as KEY=VALUE (repeatable)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	_ = cmd.MarkFlagRequired("endpoint-url")
	_ = cmd.MarkFlagRequired("auth-type")
	return cmd
}

func makeIntegrationsCreateFromCatalogCmd() *cobra.Command {
	var (
		catalogID   string
		authType    string
		credential  string
		slug        string
		description string
		project     string
		org         string
	)
	cmd := &cobra.Command{
		Use:   "create-from-catalog <name>",
		Short: "Connect an MCP server from the catalog",
		Long: `Create an integration connection from a curated catalog entry. The endpoint and
transport come from the catalog entry; you supply a name and (unless the entry
needs no auth) a credential.

Use 'iai integrations catalog' to find the --catalog-id.

The connection is verified against the live server on save.

Examples:
  iai integrations create-from-catalog github \
    --catalog-id github --auth-type bearer --credential "$GITHUB_TOKEN"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			if err := validateMcpAuth(authType, credential); err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			body := clients.McpConnectionCreateBody{
				Type:        "platform",
				CatalogID:   catalogID,
				Name:        name,
				Slug:        slug,
				Description: description,
				AuthType:    authType,
				Credential:  credential,
			}

			fmt.Fprintf(out, "\nConnecting %q from catalog entry %q and verifying...\n\n", name, catalogID)
			conn, err := apiClient.CreateMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, body)
			if err != nil {
				return err
			}
			return output.PrintMcpConnectionDetail(out, conn)
		},
	}
	cmd.Flags().StringVar(&catalogID, "catalog-id", "", "Catalog entry id (required; see 'iai integrations catalog')")
	cmd.Flags().StringVar(&authType, "auth-type", "", "Auth type: api_key, bearer, or none (required)")
	cmd.Flags().StringVar(&credential, "credential", "", "API key or bearer token (required unless auth-type=none)")
	cmd.Flags().StringVar(&slug, "slug", "", "Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)")
	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	_ = cmd.MarkFlagRequired("catalog-id")
	_ = cmd.MarkFlagRequired("auth-type")
	return cmd
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./cmd/ -run 'TestValidateMcp|TestParseHeaderFlags' -v`
Expected: PASS.

- [ ] **Step 5: Build**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go build ./... && go run main.go integrations create-custom --help`
Expected: builds; help shows the custom-create flags.

- [ ] **Step 6: Commit**

```bash
git add cmd/integrations.go cmd/integrations_test.go
git commit -m "feat(integrations): add create-custom and create-from-catalog commands"
```

---

## Task 7: `delete` and `verify`

**Files:**
- Modify: `cmd/integrations.go`

- [ ] **Step 1: Replace the delete/verify stubs with real implementations**

Update the import block to add `"bufio"`. Replace the two stubs:

```go
func makeIntegrationsDeleteCmd() *cobra.Command {
	var (
		force   bool
		project string
		org     string
	)
	cmd := &cobra.Command{
		Use:     "delete <connection-id>",
		Aliases: []string{"rm"},
		Short:   "Delete an integration connection",
		Long: `Delete an integration connection and its cached tools. This does not affect the
remote MCP server. Use -f to skip the confirmation prompt.

Examples:
  iai integrations delete 3f9c1a2e-...
  iai integrations delete 3f9c1a2e-... -f`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])

			if !force {
				fmt.Fprintf(out, "This will delete integration connection %q. Continue? [y/N] ", id)
				reader := bufio.NewReader(cmd.InOrStdin())
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Fprintln(out, "Aborted.")
					return nil
				}
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			if err := apiClient.DeleteMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id); err != nil {
				return err
			}
			fmt.Fprintf(out, "Successfully deleted integration connection %q.\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func makeIntegrationsVerifyCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:   "verify <connection-id>",
		Short: "Re-verify a connection and refresh its tools",
		Long: `Re-dial the MCP server for a connection (initialize + list tools) and refresh the
cached tool list. Reports the connection status and, on failure, the error class
and message.

Examples:
  iai integrations verify 3f9c1a2e-...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			res, err := apiClient.VerifyMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id)
			if err != nil {
				return err
			}
			return output.PrintMcpVerifyResult(out, res)
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}
```

- [ ] **Step 2: Build and smoke-test help**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go build ./... && go run main.go integrations delete --help && go run main.go integrations verify --help`
Expected: builds; both help screens render.

- [ ] **Step 3: Commit**

```bash
git add cmd/integrations.go
git commit -m "feat(integrations): add delete and verify commands"
```

---

## Task 8: `tools run` + JSON-args parsing

**Files:**
- Modify: `cmd/integrations.go`
- Modify: `cmd/integrations_test.go`

- [ ] **Step 1: Write the failing args-parsing test**

Append to `cmd/integrations_test.go`:

```go
func TestResolveToolArgs(t *testing.T) {
	// default empty
	got, err := resolveToolArgs("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %#v", got)
	}

	// inline JSON object
	got, err = resolveToolArgs(`{"q":"foo","n":2}`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["q"] != "foo" {
		t.Fatalf("unexpected args: %#v", got)
	}

	// non-object JSON rejected
	if _, err := resolveToolArgs(`[1,2,3]`, ""); err == nil {
		t.Fatal("expected error for non-object JSON")
	}

	// invalid JSON rejected
	if _, err := resolveToolArgs(`{not json}`, ""); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
```

> Note: the `--args` / `--args-file` mutual exclusion is enforced by Cobra's `MarkFlagsMutuallyExclusive`, so `resolveToolArgs` only handles the "at most one provided" case. The file-reading branch is covered by integration testing via the build smoke test; unit coverage focuses on JSON shape validation.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./cmd/ -run 'TestResolveToolArgs' -v`
Expected: compile error / FAIL — `resolveToolArgs` undefined.

- [ ] **Step 3: Implement `tools` parent, `run` leaf, and `resolveToolArgs`**

Update the import block to add `"encoding/json"` and `"os"`. Replace the `tools` stub:

```go
func resolveToolArgs(inline, file string) (map[string]any, error) {
	raw := inline
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read --args-file %q: %w", file, err)
		}
		raw = string(data)
	}
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: must be a JSON object: %w", err)
	}
	return args, nil
}

func makeIntegrationsToolsCmd() *cobra.Command {
	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "Inspect and run tools on a connection",
		Long:  `Subcommands for working with the tools exposed by an integration connection.`,
	}
	toolsCmd.AddCommand(makeIntegrationsToolsRunCmd())
	return toolsCmd
}

func makeIntegrationsToolsRunCmd() *cobra.Command {
	var (
		argsJSON string
		argsFile string
		project  string
		org      string
	)
	cmd := &cobra.Command{
		Use:   "run <connection-id> <tool-name>",
		Short: "Run a tool on a connection",
		Long: `Invoke a tool exposed by a connection's MCP server and print the result.

Only enabled, server-advertised tools can run. Arguments are a JSON object passed
inline with --args or from a file with --args-file (mutually exclusive). When
omitted, an empty argument object is sent.

Examples:
  iai integrations tools run 3f9c1a2e-... search --args '{"query":"langfuse"}'
  iai integrations tools run 3f9c1a2e-... search --args-file ./args.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])
			tool := strings.TrimSpace(args[1])

			toolArgs, err := resolveToolArgs(argsJSON, argsFile)
			if err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			res, err := apiClient.RunMcpTool(cmd.Context(), pCtx.orgId, pCtx.projectId, id, tool, toolArgs)
			if err != nil {
				return err
			}
			return output.PrintMcpToolResult(out, res)
		},
	}
	cmd.Flags().StringVar(&argsJSON, "args", "", "Tool arguments as an inline JSON object")
	cmd.Flags().StringVar(&argsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	cmd.MarkFlagsMutuallyExclusive("args", "args-file")
	return cmd
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go test ./cmd/ -run 'TestResolveToolArgs' -v`
Expected: PASS.

- [ ] **Step 5: Full build + test + help smoke**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go build ./... && go test ./... && go run main.go integrations tools run --help`
Expected: build clean; all tests PASS; help renders. Confirm no leftover stubs remain in `cmd/integrations.go`.

- [ ] **Step 6: Commit**

```bash
git add cmd/integrations.go cmd/integrations_test.go
git commit -m "feat(integrations): add tools run command"
```

---

## Task 9: Lint, regenerate docs, final verification

**Files:**
- Create: `docs/iai_integrations*.md` (generated)

- [ ] **Step 1: Run the full test suite and vet**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go vet ./... && go test ./...`
Expected: vet clean; all tests PASS.

- [ ] **Step 2: Regenerate CLI docs**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && make docs`
Expected: new files under `docs/`: `iai_integrations.md`, `iai_integrations_list.md`, `iai_integrations_get.md`, `iai_integrations_catalog.md`, `iai_integrations_create-custom.md`, `iai_integrations_create-from-catalog.md`, `iai_integrations_delete.md`, `iai_integrations_verify.md`, `iai_integrations_tools.md`, `iai_integrations_tools_run.md`. The root `iai.md` (if present) gains the `integrations` entry.

- [ ] **Step 3: Verify the generated docs**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && git status --short docs/ && ls docs/iai_integrations*`
Expected: the integrations doc files are listed as new/modified.

- [ ] **Step 4: Commit the docs**

```bash
git add docs/
git commit -m "docs(integrations): regenerate cli docs for integrations commands"
```

- [ ] **Step 5: Final end-to-end help review**

Run: `cd /Users/pedropalacioestrada/Source/interactive-cli && go run main.go integrations --help && go run main.go --help | grep -A1 integrations`
Expected: `integrations` appears under the Infrastructure group in `iai --help`; the subcommand help lists all eight commands.

---

## Self-Review Notes (author checklist — verified)

- **Spec coverage:** all 7 endpoints → client methods (Tasks 2–3); all 8 commands (Tasks 5–8); 5 printers (Task 4); docs regeneration (Task 9); tests at client/printer/validation layers. Copilot wiring correctly excluded.
- **Type consistency:** `McpConnectionCreateBody`, `McpConnectionDetail`, `McpVerifyData`, `McpToolCallData`, `McpCatalogEntry` used identically across client, output, and cmd tasks. Method names (`ListMcpConnections`, `GetMcpConnection`, `CreateMcpConnection`, `DeleteMcpConnection`, `VerifyMcpConnection`, `RunMcpTool`, `ListMcpCatalog`) match between definition (Tasks 2–3) and call sites (Tasks 5–8). Printer names (`PrintMcpConnectionList`, `PrintMcpConnectionDetail`, `PrintMcpCatalog`, `PrintMcpVerifyResult`, `PrintMcpToolResult`) match between Task 4 and call sites.
- **Auth:** MCP methods omit `requireAPIKeyMode()` by design (header note) — preserves cookie + bearer-JWT (copilot) auth.
- **Intermediate-build caveat:** Tasks 1, 5 note the import/stub ordering needed to keep incremental builds green; same-pass implementation avoids it.
