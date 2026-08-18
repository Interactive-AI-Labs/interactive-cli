package sync

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestAllowDeleteResource(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		resource string
		want     bool
	}{
		{
			name:     "nil list",
			allowed:  nil,
			resource: "databases",
			want:     false,
		},
		{
			name:     "empty list",
			allowed:  []string{},
			resource: "databases",
			want:     false,
		},
		{
			name:     "exact match",
			allowed:  []string{"databases"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "case insensitive match",
			allowed:  []string{"Databases"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "all keyword",
			allowed:  []string{"all"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "ALL keyword uppercase",
			allowed:  []string{"ALL"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "no match",
			allowed:  []string{"services"},
			resource: "databases",
			want:     false,
		},
		{
			name:     "multiple entries with match",
			allowed:  []string{"services", "databases"},
			resource: "databases",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllowDeleteResource(tt.allowed, tt.resource)
			if got != tt.want {
				t.Errorf(
					"AllowDeleteResource(%v, %q) = %v, want %v",
					tt.allowed, tt.resource, got, tt.want,
				)
			}
		})
	}
}

func TestMcpsRejectsCredentialedUpdateWithoutCredential(t *testing.T) {
	tests := []struct {
		name string
		body clients.CreateMcpBody
	}{
		{name: "auth omitted", body: clients.CreateMcpBody{}},
		{
			name: "credential omitted",
			body: clients.CreateMcpBody{Auth: clients.McpAuthBody{Type: "bearer"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations/o1/projects/p1/mcps":
					fmt.Fprint(
						w,
						`{"mcps":[{"name":"tools","projectId":"p1","revision":1,"type":"external","auth":{"type":"bearer"}}]}`,
					)
				case r.Method == http.MethodPut:
					t.Errorf("credentialed mcp update reached PUT without auth.credential")
					fmt.Fprint(w, `{}`)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			})

			_, err := Mcps(
				context.Background(),
				&bytes.Buffer{},
				client,
				"o1",
				"p1",
				"stack-1",
				map[string]clients.CreateMcpBody{"tools": tt.body},
				Options{},
			)
			if err == nil || !strings.Contains(err.Error(), "auth.credential is required") {
				t.Fatalf("Mcps() error = %v, want auth.credential error", err)
			}
		})
	}
}

func TestPrintResult(t *testing.T) {
	tests := []struct {
		name    string
		label   string
		result  *Result
		syncErr error
		want    string
		wantErr bool
	}{
		{
			name:  "created and deleted",
			label: "databases",
			result: &Result{
				Created: []string{"new-db"},
				Deleted: []string{"old-db"},
			},
			want: "Created databases: new-db\n" +
				"Deleted databases: old-db\n",
		},
		{
			name:  "updated items",
			label: "services",
			result: &Result{
				Updated: []string{"svc-a"},
			},
			want: "Updated services: svc-a\n",
		},
		{
			name:   "no changes",
			label:  "services",
			result: &Result{},
			want:   "No changes required; services already match config.\n",
		},
		{
			name:  "multiple items joined with comma",
			label: "services",
			result: &Result{
				Created: []string{"svc-a", "svc-b"},
				Deleted: []string{"svc-c", "svc-d"},
			},
			want: "Created services: svc-a, svc-b\n" +
				"Deleted services: svc-c, svc-d\n",
		},
		{
			name:  "protected items print warning",
			label: "databases",
			result: &Result{
				Created:   []string{"new-db"},
				Protected: []string{"old-db"},
			},
			want: "Created databases: new-db\n" +
				"\nProtected databases (not deleted): old-db\n" +
				"Use --allow-delete=databases to delete them.\n",
		},
		{
			name:    "error with partial result",
			label:   "services",
			result:  &Result{Created: []string{"svc-a"}},
			syncErr: fmt.Errorf("failed to create service \"svc-b\""),
			wantErr: true,
			want:    "Created services (partial): svc-a\n",
		},
		{
			name:    "error with nil result",
			label:   "services",
			syncErr: fmt.Errorf("failed to list services"),
			wantErr: true,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintResult(&buf, tt.label, tt.result, tt.syncErr)
			if tt.wantErr && err != tt.syncErr {
				t.Fatalf("expected original error, got: %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintPlan(t *testing.T) {
	tests := []struct {
		name   string
		label  string
		result *Result
		want   string
	}{
		{
			name:  "full plan",
			label: "services",
			result: &Result{
				Created:   []string{"svc-new"},
				Updated:   []string{"svc-a", "svc-b"},
				Deleted:   []string{"svc-gone"},
				Protected: []string{"svc-old"},
			},
			want: "Would create services: svc-new\n" +
				"Would update services: svc-a, svc-b\n" +
				"Would delete services: svc-gone\n" +
				"Would refuse to delete services: svc-old (a config that omits a resource looks identical to a stale one — pass --allow-delete=services to delete)\n",
		},
		{
			name:   "no changes",
			label:  "agents",
			result: &Result{},
			want:   "No changes required; agents already match config.\n",
		},
		{
			name:  "only refused deletions",
			label: "databases",
			result: &Result{
				Protected: []string{"old-db"},
			},
			want: "Would refuse to delete databases: old-db (a config that omits a resource looks identical to a stale one — pass --allow-delete=databases to delete)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintPlan(&buf, tt.label, tt.result)
			if got := buf.String(); got != tt.want {
				t.Errorf("plan mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func newTestDeployClient(t *testing.T, handler http.HandlerFunc) *clients.DeploymentClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := clients.NewDeploymentClient(server.URL, 5*time.Second, "test-token", "", nil)
	if err != nil {
		t.Fatalf("NewDeploymentClient() error = %v", err)
	}
	return client
}

func TestServicesPrintsUpdateBanner(t *testing.T) {
	client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations/o1/projects/p1/services":
			fmt.Fprint(
				w,
				`{"services":[{"name":"svc-a","projectId":"p1","revision":3,"status":"ready","updated":"2026-07-24T11:20:00Z"}]}`,
			)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/organizations/o1/projects/p1/services/svc-a":
			fmt.Fprint(w, `{}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/organizations/o1/projects/p1/services/svc-new":
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]clients.CreateServiceBody{
		"svc-a":   {},
		"svc-new": {},
	}
	result, err := Services(
		context.Background(),
		&warn,
		client,
		"o1",
		"p1",
		"stack-1",
		desired,
		Options{},
	)
	if err != nil {
		t.Fatalf("Services() error = %v", err)
	}

	wantWarn := "Live: service svc-a revision 3, last updated 2026-07-24 11:20 UTC — this update creates revision 4\n"
	if got := warn.String(); got != wantWarn {
		t.Errorf("banner = %q, want %q", got, wantWarn)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "svc-a" {
		t.Errorf("Updated = %v, want [svc-a]", result.Updated)
	}
	if len(result.Created) != 1 || result.Created[0] != "svc-new" {
		t.Errorf("Created = %v, want [svc-new]", result.Created)
	}
}

func TestSyncRefusesDeletionsByDefault(t *testing.T) {
	tests := []struct {
		name          string
		listPath      string
		listBody      string
		deletePath    string
		putPath       string
		wantWarn      string
		wantProtected []string
		wantUpdated   []string
		run           func(context.Context, *bytes.Buffer, *clients.DeploymentClient) (*Result, error)
	}{
		{
			name:       "services",
			listPath:   "/v1/organizations/o1/projects/p1/services",
			listBody:   `{"services":[{"name":"svc-a","projectId":"p1","revision":3,"status":"ready","updated":"2026-07-24T11:20:00Z"},{"name":"svc-old","projectId":"p1","revision":9,"status":"ready"}]}`,
			deletePath: "/v1/organizations/o1/projects/p1/services/svc-old",
			putPath:    "/v1/organizations/o1/projects/p1/services/svc-a",
			wantWarn: "⚠ sync will NOT delete 1 service not in the config: svc-old" +
				" (a config that omits a resource looks identical to a stale one — pass --allow-delete=services to delete)\n" +
				"Live: service svc-a revision 3, last updated 2026-07-24 11:20 UTC — this update creates revision 4\n",
			wantProtected: []string{"svc-old"},
			wantUpdated:   []string{"svc-a"},
			run: func(ctx context.Context, warn *bytes.Buffer, client *clients.DeploymentClient) (*Result, error) {
				return Services(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]clients.CreateServiceBody{"svc-a": {}}, Options{})
			},
		},
		{
			name:       "agents",
			listPath:   "/v1/organizations/o1/projects/p1/agents",
			listBody:   `{"agents":[{"name":"agent-old","projectId":"p1","revision":5,"status":"ready"}]}`,
			deletePath: "/v1/organizations/o1/projects/p1/agents/agent-old",
			wantWarn: "⚠ sync will NOT delete 1 agent not in the config: agent-old" +
				" (a config that omits a resource looks identical to a stale one — pass --allow-delete=agents to delete)\n",
			wantProtected: []string{"agent-old"},
			run: func(ctx context.Context, warn *bytes.Buffer, client *clients.DeploymentClient) (*Result, error) {
				return Agents(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]clients.CreateAgentBody{}, Options{})
			},
		},
		{
			name:       "databases",
			listPath:   "/v1/organizations/o1/projects/p1/databases",
			listBody:   `{"databases":[{"name":"old-db","revision":2,"status":"ready"}]}`,
			deletePath: "/v1/organizations/o1/projects/p1/databases/old-db",
			wantWarn: "⚠ sync will NOT delete 1 database not in the config: old-db" +
				" (a config that omits a resource looks identical to a stale one — pass --allow-delete=databases to delete)\n",
			wantProtected: []string{"old-db"},
			run: func(ctx context.Context, warn *bytes.Buffer, client *clients.DeploymentClient) (*Result, error) {
				return Databases(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]clients.CreateDatabaseBody{}, Options{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == tt.listPath:
					fmt.Fprint(w, tt.listBody)
				case tt.putPath != "" && r.Method == http.MethodPut && r.URL.Path == tt.putPath:
					fmt.Fprint(w, `{}`)
				case r.Method == http.MethodDelete && r.URL.Path == tt.deletePath:
					t.Errorf("%s was deleted without --allow-delete", tt.deletePath)
					fmt.Fprint(w, `{}`)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			})

			var warn bytes.Buffer
			result, err := tt.run(context.Background(), &warn, client)
			if err != nil {
				t.Fatalf("sync error = %v", err)
			}
			if got := warn.String(); got != tt.wantWarn {
				t.Errorf("warnings = %q, want %q", got, tt.wantWarn)
			}
			if len(result.Deleted) != 0 {
				t.Errorf("Deleted = %v, want []", result.Deleted)
			}
			if len(result.Protected) != len(tt.wantProtected) ||
				(len(tt.wantProtected) > 0 && result.Protected[0] != tt.wantProtected[0]) {
				t.Errorf("Protected = %v, want %v", result.Protected, tt.wantProtected)
			}
			if len(result.Updated) != len(tt.wantUpdated) ||
				(len(tt.wantUpdated) > 0 && result.Updated[0] != tt.wantUpdated[0]) {
				t.Errorf("Updated = %v, want %v", result.Updated, tt.wantUpdated)
			}
		})
	}
}

func TestSyncDeletesWithAllowDelete(t *testing.T) {
	tests := []struct {
		name        string
		listPath    string
		listBody    string
		deletePath  string
		putPath     string
		wantWarn    string
		wantDeleted []string
		run         func(context.Context, *bytes.Buffer, *clients.DeploymentClient) (*Result, error)
	}{
		{
			name:       "services",
			listPath:   "/v1/organizations/o1/projects/p1/services",
			listBody:   `{"services":[{"name":"svc-a","projectId":"p1","revision":3,"status":"ready","updated":"2026-07-24T11:20:00Z"},{"name":"svc-old","projectId":"p1","revision":9,"status":"ready"}]}`,
			deletePath: "/v1/organizations/o1/projects/p1/services/svc-old",
			putPath:    "/v1/organizations/o1/projects/p1/services/svc-a",
			wantWarn: "⚠ sync will DELETE 1 service not in the config: svc-old" +
				" (--allow-delete=services; service deletes run after service creates/updates)\n" +
				"Live: service svc-a revision 3, last updated 2026-07-24 11:20 UTC — this update creates revision 4\n",
			wantDeleted: []string{"svc-old"},
			run: func(ctx context.Context, warn *bytes.Buffer, client *clients.DeploymentClient) (*Result, error) {
				return Services(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]clients.CreateServiceBody{"svc-a": {}}, Options{AllowDelete: true})
			},
		},
		{
			name:       "agents",
			listPath:   "/v1/organizations/o1/projects/p1/agents",
			listBody:   `{"agents":[{"name":"agent-old","projectId":"p1","revision":5,"status":"ready"}]}`,
			deletePath: "/v1/organizations/o1/projects/p1/agents/agent-old",
			wantWarn: "⚠ sync will DELETE 1 agent not in the config: agent-old" +
				" (--allow-delete=agents; agent deletes run after agent creates/updates)\n",
			wantDeleted: []string{"agent-old"},
			run: func(ctx context.Context, warn *bytes.Buffer, client *clients.DeploymentClient) (*Result, error) {
				return Agents(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]clients.CreateAgentBody{}, Options{AllowDelete: true})
			},
		},
		{
			name:       "databases",
			listPath:   "/v1/organizations/o1/projects/p1/databases",
			listBody:   `{"databases":[{"name":"old-db","revision":2,"status":"ready"}]}`,
			deletePath: "/v1/organizations/o1/projects/p1/databases/old-db",
			wantWarn: "⚠ sync will DELETE 1 database not in the config: old-db" +
				" (--allow-delete=databases; database deletes run after database creates/updates)\n",
			wantDeleted: []string{"old-db"},
			run: func(ctx context.Context, warn *bytes.Buffer, client *clients.DeploymentClient) (*Result, error) {
				return Databases(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]clients.CreateDatabaseBody{}, Options{AllowDelete: true})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deleted bool
			client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == tt.listPath:
					fmt.Fprint(w, tt.listBody)
				case tt.putPath != "" && r.Method == http.MethodPut && r.URL.Path == tt.putPath:
					fmt.Fprint(w, `{}`)
				case r.Method == http.MethodDelete && r.URL.Path == tt.deletePath:
					deleted = true
					fmt.Fprint(w, `{}`)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			})

			var warn bytes.Buffer
			result, err := tt.run(context.Background(), &warn, client)
			if err != nil {
				t.Fatalf("sync error = %v", err)
			}
			if got := warn.String(); got != tt.wantWarn {
				t.Errorf("warnings = %q, want %q", got, tt.wantWarn)
			}
			if !deleted {
				t.Errorf("%s was announced but not deleted", tt.deletePath)
			}
			if len(result.Deleted) != len(tt.wantDeleted) ||
				(len(tt.wantDeleted) > 0 && result.Deleted[0] != tt.wantDeleted[0]) {
				t.Errorf("Deleted = %v, want %v", result.Deleted, tt.wantDeleted)
			}
			if len(result.Protected) != 0 {
				t.Errorf("Protected = %v, want []", result.Protected)
			}
		})
	}
}

func TestServicesDryRunPlansWithoutWriting(t *testing.T) {
	client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations/o1/projects/p1/services":
			fmt.Fprint(
				w,
				`{"services":[{"name":"svc-a","projectId":"p1","revision":3,"status":"ready","updated":"2026-07-24T11:20:00Z"},{"name":"svc-old","projectId":"p1","revision":9,"status":"ready"}]}`,
			)
		default:
			t.Errorf("dry run made a write: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]clients.CreateServiceBody{
		"svc-a":   {},
		"svc-new": {},
	}
	result, err := Services(
		context.Background(), &warn, client, "o1", "p1", "stack-1", desired,
		Options{DryRun: true},
	)
	if err != nil {
		t.Fatalf("Services() error = %v", err)
	}

	if got := warn.String(); got != "" {
		t.Errorf("dry run printed warnings = %q, want none (the plan is the deliverable)", got)
	}
	if len(result.Created) != 1 || result.Created[0] != "svc-new" {
		t.Errorf("Created = %v, want [svc-new]", result.Created)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "svc-a" {
		t.Errorf("Updated = %v, want [svc-a]", result.Updated)
	}
	if len(result.Protected) != 1 || result.Protected[0] != "svc-old" {
		t.Errorf("Protected = %v, want [svc-old]", result.Protected)
	}
	if len(result.Deleted) != 0 {
		t.Errorf("Deleted = %v, want []", result.Deleted)
	}
}

func TestAgentsPrintsUpdateBanner(t *testing.T) {
	client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations/o1/projects/p1/agents":
			fmt.Fprint(
				w,
				`{"agents":[{"name":"agent-a","projectId":"p1","revision":13,"status":"ready","updated":"2026-07-24T11:20:00Z"}]}`,
			)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/organizations/o1/projects/p1/agents/agent-a":
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]clients.CreateAgentBody{"agent-a": {}}
	result, err := Agents(
		context.Background(),
		&warn,
		client,
		"o1",
		"p1",
		"stack-1",
		desired,
		Options{},
	)
	if err != nil {
		t.Fatalf("Agents() error = %v", err)
	}

	wantWarn := "Live: agent agent-a revision 13, last updated 2026-07-24 11:20 UTC — this update creates revision 14\n"
	if got := warn.String(); got != wantWarn {
		t.Errorf("banner = %q, want %q", got, wantWarn)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "agent-a" {
		t.Errorf("Updated = %v, want [agent-a]", result.Updated)
	}
}
