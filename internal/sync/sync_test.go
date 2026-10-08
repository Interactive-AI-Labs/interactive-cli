package sync

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
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

func TestRequireMcpCredential(t *testing.T) {
	tests := []struct {
		name       string
		authType   string
		credential string
		wantErr    bool
	}{
		{name: "unknown auth type passes"},
		{name: "none needs no credential", authType: "none"},
		{name: "none is matched case-insensitively", authType: "None"},
		{name: "bearer without credential is refused", authType: "bearer", wantErr: true},
		{name: "api_key without credential is refused", authType: "api_key", wantErr: true},
		{name: "bearer with credential passes", authType: "bearer", credential: "token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireMcpCredential("tools", tt.authType, tt.credential)
			if (err != nil) != tt.wantErr {
				t.Fatalf("requireMcpCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "auth.credential is required") {
				t.Fatalf("requireMcpCredential() error = %v, want auth.credential error", err)
			}
		})
	}
}

func TestMcpConflicts(t *testing.T) {
	stack := func(s string) *string { return &s }
	operator := map[string]deployment.McpOutput{
		"tools":     {Name: "tools", Type: "internal", StackId: "s1"},
		"legacy":    {Name: "legacy", Type: "external", StackId: "s1"},
		"svc":       {Name: "svc", Type: "internal", StackId: "s2"},
		"oldremote": {Name: "oldremote", Type: "external", StackId: "s2"},
	}
	external := platform.McpBackendExternal
	platformMcps := map[string]platform.McpSchema{
		"tools": {Name: "tools", Backend: platform.McpBackendInternal, StackId: stack("s1")},
		"svc":   {Name: "svc", Backend: platform.McpBackendInternal, StackId: stack("s2")},
		"docs":  {Name: "docs", Backend: external, StackId: stack("s1"), AuthType: stack("none")},
		"github": {
			Name:     "github",
			Backend:  external,
			StackId:  stack("s1"),
			AuthType: stack("oauth"),
		},
		"shared": {
			Name:     "shared",
			Backend:  external,
			StackId:  stack("s2"),
			AuthType: stack("bearer"),
		},
		"loose": {Name: "loose", Backend: external, AuthType: stack("none")},
	}
	tests := []struct {
		name            string
		selfHosted      []string
		remote          []string
		wantTypeChanged []string
		wantUnmanaged   []string
	}{
		{
			name:       "same types and stack on both sides",
			selfHosted: []string{"tools"},
			remote:     []string{"docs"},
		},
		{
			name:            "self-hosted mcp that is remote on the platform",
			selfHosted:      []string{"docs"},
			wantTypeChanged: []string{"docs"},
		},
		{
			name:          "self-hosted mcp that is remote in another stack",
			selfHosted:    []string{"shared"},
			wantUnmanaged: []string{"shared"},
		},
		{
			name:          "self-hosted mcp that is remote without a stack",
			selfHosted:    []string{"loose"},
			wantUnmanaged: []string{"loose"},
		},
		{
			name:          "self-hosted mcp that is self-hosted in another stack",
			selfHosted:    []string{"svc"},
			wantUnmanaged: []string{"svc"},
		},
		{
			name:            "self-hosted mcp still remote on the operator",
			selfHosted:      []string{"legacy"},
			wantTypeChanged: []string{"legacy"},
		},
		{
			name:          "self-hosted mcp named like an oauth remote mcp",
			selfHosted:    []string{"github"},
			wantUnmanaged: []string{"github"},
		},
		{
			name:            "remote mcp that is self-hosted on the operator",
			remote:          []string{"tools"},
			wantTypeChanged: []string{"tools"},
		},
		{
			name:          "remote mcp that is self-hosted in another stack",
			remote:        []string{"svc"},
			wantUnmanaged: []string{"svc"},
		},
		{
			name:            "remote mcp still on the operator",
			remote:          []string{"legacy"},
			wantTypeChanged: []string{"legacy"},
		},
		{
			name:          "remote mcp still on the operator in another stack",
			remote:        []string{"oldremote"},
			wantUnmanaged: []string{"oldremote"},
		},
		{
			name:          "remote mcp owned by another stack",
			remote:        []string{"shared"},
			wantUnmanaged: []string{"shared"},
		},
		{
			name:          "remote mcp without a stack",
			remote:        []string{"loose"},
			wantUnmanaged: []string{"loose"},
		},
		{
			name:          "remote mcp set up with iai mcps",
			remote:        []string{"github"},
			wantUnmanaged: []string{"github"},
		},
		{
			name:       "new names conflict with nothing",
			selfHosted: []string{"new"},
			remote:     []string{"other"},
		},
		{
			name:       "conflicts are sorted",
			selfHosted: []string{"docs"},
			remote: []string{
				"tools",
				"legacy",
				"shared",
				"loose",
				"svc",
				"github",
				"oldremote",
			},
			wantTypeChanged: []string{"docs", "legacy", "tools"},
			wantUnmanaged:   []string{"github", "loose", "oldremote", "shared", "svc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selfHosted := make(map[string]deployment.CreateMcpBody, len(tt.selfHosted))
			for _, name := range tt.selfHosted {
				selfHosted[name] = deployment.CreateMcpBody{}
			}
			remote := make(map[string]platform.McpCreateRequest, len(tt.remote))
			for _, name := range tt.remote {
				remote[name] = platform.McpCreateRequest{Name: name}
			}
			typeChanged, unmanaged := mcpConflicts(
				operator,
				platformMcps,
				"s1",
				selfHosted,
				remote,
			)
			if !reflect.DeepEqual(typeChanged, tt.wantTypeChanged) ||
				!reflect.DeepEqual(unmanaged, tt.wantUnmanaged) {
				t.Fatalf(
					"mcpConflicts() = %v, %v, want %v, %v",
					typeChanged, unmanaged, tt.wantTypeChanged, tt.wantUnmanaged,
				)
			}
		})
	}
}

func TestRemoteMcpPatch(t *testing.T) {
	str := func(s string) *string { return &s }
	none := platform.McpAuth{Type: platform.McpAuthNone}
	bearer := func(credential string) platform.McpAuth {
		auth := platform.McpAuth{Type: platform.McpAuthBearer}
		if credential != "" {
			auth.Credential = str(credential)
		}
		return auth
	}
	liveGithub := platform.McpSchema{
		CatalogID: str("github"), AuthType: str("bearer"), HasCredential: true,
	}
	liveAcme := func(authType string) platform.McpSchema {
		return platform.McpSchema{
			EndpointURL: str(
				"https://mcp.acme.com/mcp",
			),
			AuthType:      str(authType),
			HasCredential: true,
		}
	}
	tests := []struct {
		name    string
		live    platform.McpSchema
		body    platform.McpCreateRequest
		want    platform.McpUpdateRequest
		wantErr string
	}{
		{
			name: "catalog entry without credential",
			live: platform.McpSchema{CatalogID: str("awsknowledge"), AuthType: str("none")},
			body: platform.McpCreateRequest{CatalogID: str("awsknowledge"), Auth: none},
			want: platform.McpUpdateRequest{},
		},
		{
			name: "credential rotation on a catalog entry",
			live: liveGithub,
			body: platform.McpCreateRequest{CatalogID: str("github"), Auth: bearer("rotated")},
			want: platform.McpUpdateRequest{
				"auth": map[string]any{"type": "bearer", "credential": "rotated"},
			},
		},
		{
			name: "no credential in the config keeps the live one",
			live: liveGithub,
			body: platform.McpCreateRequest{CatalogID: str("github"), Auth: bearer("")},
			want: platform.McpUpdateRequest{},
		},
		{
			name:    "no credential in the config and none live",
			live:    platform.McpSchema{CatalogID: str("github"), AuthType: str("bearer")},
			body:    platform.McpCreateRequest{CatalogID: str("github"), Auth: bearer("")},
			wantErr: "auth.credential is required",
		},
		{
			name:    "no credential with an auth type change",
			live:    platform.McpSchema{CatalogID: str("github"), AuthType: str("none")},
			body:    platform.McpCreateRequest{CatalogID: str("github"), Auth: bearer("")},
			wantErr: "auth.credential is required",
		},
		{
			name: "custom always needs its credential",
			live: liveAcme("custom"),
			body: platform.McpCreateRequest{
				EndpointURL: str("https://mcp.acme.com/mcp"),
				Auth: platform.McpAuth{
					Type:       platform.McpAuthCustom,
					HeaderName: str("X-Token"),
				},
			},
			wantErr: "auth.credential is required",
		},
		{
			name: "endpoint with custom header auth",
			live: liveAcme("custom"),
			body: platform.McpCreateRequest{
				EndpointURL: str("https://mcp.acme.com/mcp"),
				Auth: platform.McpAuth{
					Type:         platform.McpAuthCustom,
					Credential:   str("token"),
					HeaderName:   str("X-Token"),
					HeaderPrefix: str("Token "),
				},
			},
			want: platform.McpUpdateRequest{
				"auth": map[string]any{
					"type":          "custom",
					"credential":    "token",
					"header_name":   "X-Token",
					"header_prefix": "Token ",
				},
			},
		},
		{
			name: "endpoint without a header override clears a previous one",
			live: liveAcme("bearer"),
			body: platform.McpCreateRequest{
				EndpointURL: str("https://mcp.acme.com/mcp"),
				Auth:        bearer("token"),
			},
			want: platform.McpUpdateRequest{
				"auth": map[string]any{
					"type":          "bearer",
					"credential":    "token",
					"header_name":   nil,
					"header_prefix": nil,
				},
			},
		},
		{
			name: "endpoint url change without a credential keeps the live one",
			live: liveAcme("bearer"),
			body: platform.McpCreateRequest{
				EndpointURL: str("https://new.example.com/mcp"),
				Auth:        bearer(""),
			},
			want: platform.McpUpdateRequest{"endpoint_url": "https://new.example.com/mcp"},
		},
		{
			name:    "catalog entry change",
			live:    platform.McpSchema{CatalogID: str("github"), AuthType: str("none")},
			body:    platform.McpCreateRequest{CatalogID: str("gitlab"), Auth: none},
			wantErr: "changed its catalog entry",
		},
		{
			name: "catalog entry to endpoint url",
			live: platform.McpSchema{
				CatalogID:   str("github"),
				EndpointURL: str("https://api.github.com/mcp"),
				AuthType:    str("none"),
			},
			body: platform.McpCreateRequest{
				EndpointURL: str("https://api.github.com/mcp"),
				Auth:        none,
			},
			wantErr: "changed its catalog entry",
		},
		{
			name: "endpoint url to catalog entry",
			live: platform.McpSchema{
				EndpointURL: str("https://api.github.com/mcp"),
				AuthType:    str("none"),
			},
			body:    platform.McpCreateRequest{CatalogID: str("github"), Auth: none},
			wantErr: "changed its catalog entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := remoteMcpPatch("docs", tt.live, tt.body)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("remoteMcpPatch() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("remoteMcpPatch() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("remoteMcpPatch() = %#v, want %#v", got, tt.want)
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

func newTestDeployClient(t *testing.T, handler http.HandlerFunc) *deployment.DeploymentClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := deployment.NewDeploymentClient(server.URL, 5*time.Second, "test-token", "", nil)
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
			fmt.Fprint(w, `{"changed":true}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/organizations/o1/projects/p1/services/svc-new":
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]deployment.CreateServiceBody{
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
		run           func(context.Context, *bytes.Buffer, *deployment.DeploymentClient) (*Result, error)
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
			run: func(ctx context.Context, warn *bytes.Buffer, client *deployment.DeploymentClient) (*Result, error) {
				return Services(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]deployment.CreateServiceBody{"svc-a": {}}, Options{})
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
			run: func(ctx context.Context, warn *bytes.Buffer, client *deployment.DeploymentClient) (*Result, error) {
				return Agents(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]deployment.CreateAgentBody{}, Options{})
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
			run: func(ctx context.Context, warn *bytes.Buffer, client *deployment.DeploymentClient) (*Result, error) {
				return Databases(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]deployment.CreateDatabaseBody{}, Options{})
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
					fmt.Fprint(w, `{"changed":true}`)
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
		run         func(context.Context, *bytes.Buffer, *deployment.DeploymentClient) (*Result, error)
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
			run: func(ctx context.Context, warn *bytes.Buffer, client *deployment.DeploymentClient) (*Result, error) {
				return Services(
					ctx,
					warn,
					client,
					"o1",
					"p1",
					"stack-1",
					map[string]deployment.CreateServiceBody{
						"svc-a": {},
					},
					Options{AllowDelete: true},
				)
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
			run: func(ctx context.Context, warn *bytes.Buffer, client *deployment.DeploymentClient) (*Result, error) {
				return Agents(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]deployment.CreateAgentBody{}, Options{AllowDelete: true})
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
			run: func(ctx context.Context, warn *bytes.Buffer, client *deployment.DeploymentClient) (*Result, error) {
				return Databases(ctx, warn, client, "o1", "p1", "stack-1",
					map[string]deployment.CreateDatabaseBody{}, Options{AllowDelete: true})
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
					fmt.Fprint(w, `{"changed":true}`)
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
		case r.Method == http.MethodPut && r.URL.Path == "/v1/organizations/o1/projects/p1/services/svc-a" &&
			r.URL.Query().Get("dryRun") == "true":
			fmt.Fprint(w, `{"changed":true,"servicePort":8080}`)
		default:
			t.Errorf("dry run made a write: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]deployment.CreateServiceBody{
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
			fmt.Fprint(w, `{"changed":true}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]deployment.CreateAgentBody{"agent-a": {}}
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

func TestWaitForMcps(t *testing.T) {
	type poll struct {
		state deployment.McpVerifyState
		err   error
	}
	ok := poll{state: deployment.McpVerifyState{Status: "ok"}}
	pending := poll{state: deployment.McpVerifyState{Status: "pending"}}

	tests := []struct {
		name     string
		polls    map[string][]poll // answers per fetch; the last one repeats
		order    []string
		timeout  time.Duration
		canceled bool
		wantErr  string
	}{
		{
			name:  "all ok on first poll",
			polls: map[string][]poll{"a": {ok}, "b": {ok}},
			order: []string{"a", "b"},
		},
		{
			name:  "no mcps",
			order: nil,
		},
		{
			name:  "pending then ok",
			polls: map[string][]poll{"a": {pending, pending, ok}, "b": {ok}},
			order: []string{"a", "b"},
		},
		{
			name: "error then ok",
			polls: map[string][]poll{"a": {
				{state: deployment.McpVerifyState{Status: "error", Error: "dial failed"}},
				ok,
			}},
			order: []string{"a"},
		},
		{
			name:  "failed check then ok",
			polls: map[string][]poll{"a": {{err: errors.New("bad gateway")}, ok}},
			order: []string{"a"},
		},
		{
			name:  "ok then pending waits for both at once",
			polls: map[string][]poll{"a": {ok, pending, ok}, "b": {pending, ok}},
			order: []string{"a", "b"},
		},
		{
			name:    "stays pending",
			polls:   map[string][]poll{"a": {ok}, "b": {pending}},
			order:   []string{"a", "b"},
			wantErr: "mcps not ready after 50ms: b (pending)",
		},
		{
			name: "stays error",
			polls: map[string][]poll{"a": {
				{state: deployment.McpVerifyState{Status: "error", Error: "dial failed"}},
			}},
			order:   []string{"a"},
			wantErr: "mcps not ready after 50ms: a (error: dial failed)",
		},
		{
			name:    "error without message",
			polls:   map[string][]poll{"a": {{state: deployment.McpVerifyState{Status: "error"}}}},
			order:   []string{"a"},
			wantErr: "mcps not ready after 50ms: a (error)",
		},
		{
			name:    "empty status",
			polls:   map[string][]poll{"a": {{}}},
			order:   []string{"a"},
			wantErr: "mcps not ready after 50ms: a (unknown)",
		},
		{
			name:    "check keeps failing",
			polls:   map[string][]poll{"a": {{err: errors.New("mcp not found")}}},
			order:   []string{"a"},
			wantErr: "mcps not ready after 50ms: a (check failed: mcp not found)",
		},
		{
			name:    "status is case sensitive",
			polls:   map[string][]poll{"a": {{state: deployment.McpVerifyState{Status: "OK"}}}},
			order:   []string{"a"},
			wantErr: "mcps not ready after 50ms: a (OK)",
		},
		{
			name: "reasons keep names order",
			polls: map[string][]poll{
				"a": {{state: deployment.McpVerifyState{Status: "error", Error: "timeout"}}},
				"b": {ok},
				"c": {pending},
			},
			order:   []string{"c", "a", "b"},
			wantErr: "mcps not ready after 50ms: c (pending), a (error: timeout)",
		},
		{
			name:     "canceled while waiting",
			polls:    map[string][]poll{"a": {pending}},
			order:    []string{"a"},
			timeout:  time.Hour,
			canceled: true,
			wantErr:  "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.canceled {
				cancel()
			}

			fetches := make(map[string]int)
			fetch := func(_ context.Context, name string) (deployment.McpVerifyState, error) {
				answers := tt.polls[name]
				p := answers[min(fetches[name], len(answers)-1)]
				fetches[name]++
				return p.state, p.err
			}

			timeout := cmp.Or(tt.timeout, 50*time.Millisecond)
			err := waitForMcps(ctx, fetch, tt.order, timeout, time.Millisecond)

			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("waitForMcps() error = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}
