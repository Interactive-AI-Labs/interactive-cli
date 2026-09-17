package platform

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeMcpCatalog(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantID  string
		wantErr bool
	}{
		{
			name:   "valid catalog",
			body:   `{"success":true,"data":{"entries":[{"id":"e1","name":"GitHub","category":"dev","type":"platform","auth_methods":["api_key"]}]}}`,
			wantID: "e1",
		},
		{name: "unsuccessful envelope", body: `{"success":false}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := decodeSuccess[McpCatalogListData]([]byte(tt.body), "list mcp catalog")
			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeSuccess() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(data.Entries) != 1 || data.Entries[0].ID != tt.wantID {
				t.Fatalf("entries = %#v, want ID %q", data.Entries, tt.wantID)
			}
		})
	}
}

func TestDecodeMcpToolCallResult(t *testing.T) {
	errorClass := "unauthorized"
	tests := []struct {
		name           string
		body           string
		wantStatus     string
		wantErrorClass *string
		wantResult     bool
	}{
		{
			name:           "error status carries a class and no result",
			body:           `{"success":true,"data":{"name":"socket","tool":"depscore","status":"error","error_class":"unauthorized"}}`,
			wantStatus:     "error",
			wantErrorClass: &errorClass,
		},
		{
			name:       "ok status carries a result and no class",
			body:       `{"success":true,"data":{"name":"socket","tool":"depscore","status":"ok","result":{"ok":true}}}`,
			wantStatus: "ok",
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := decodeSuccess[McpToolCallData]([]byte(tt.body), "run mcp tool")
			if err != nil {
				t.Fatalf("decodeSuccess() error = %v", err)
			}
			if diff := cmp.Diff(tt.wantStatus, res.Status); diff != "" {
				t.Errorf("status mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantErrorClass, res.ErrorClass); diff != "" {
				t.Errorf("error class mismatch (-want +got):\n%s", diff)
			}
			if hasResult := len(res.Result) > 0; hasResult != tt.wantResult {
				t.Errorf("has result = %v, want %v", hasResult, tt.wantResult)
			}
		})
	}
}

func TestMcpPath(t *testing.T) {
	tests := []struct {
		name      string
		orgID     string
		projectID string
		mcpName   string
		want      string
	}{
		{
			name: "collection path", orgID: "org-1", projectID: "project-1", mcpName: "",
			want: "/api/platform/v1/organizations/org-1/projects/project-1/mcps",
		},
		{
			name: "one mcp by name", orgID: "org-1", projectID: "project-1", mcpName: "demo",
			want: "/api/platform/v1/organizations/org-1/projects/project-1/mcps/demo",
		},
		{
			name: "a name needing escaping is escaped", orgID: "org-1", projectID: "project-1",
			mcpName: "a b/c",
			want:    "/api/platform/v1/organizations/org-1/projects/project-1/mcps/a%20b%2Fc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &APIClient{}
			if got := client.mcpPath(tt.orgID, tt.projectID, tt.mcpName); got != tt.want {
				t.Errorf("mcpPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMcpCreateRequestJSON(t *testing.T) {
	credential, endpoint := "token", "https://example.com/mcp"
	header, prefix := "X-Token", "Token "
	tests := []struct {
		name    string
		request McpCreateRequest
		want    string
	}{
		{
			name: "internal create keeps flat image",
			request: McpCreateRequest{
				Name: "demo", Backend: McpBackendInternal, Transport: "streamable_http",
				Auth: McpAuth{Type: "none"},
				Workload: &McpCreateWorkload{
					Image: "tools:2", Port: 3000, Path: "/mcp", Memory: "128M", CPU: "100m",
				},
			},
			want: `{"name":"demo","backend":"internal","transport":"streamable_http",` +
				`"auth":{"type":"none"},"workload":{"image":"tools:2","port":3000,` +
				`"endpoint":false,"path":"/mcp","memory":"128M","cpu":"100m"}}`,
		},
		{
			name: "external create carries endpoint and auth",
			request: McpCreateRequest{
				Name: "demo", Backend: McpBackendExternal,
				Transport: "streamable_http", EndpointURL: &endpoint,
				Auth: McpAuth{Type: "bearer", Credential: &credential},
			},
			want: `{"name":"demo","backend":"external",` +
				`"endpoint_url":"https://example.com/mcp","transport":"streamable_http",` +
				`"auth":{"type":"bearer","credential":"token"}}`,
		},
		{
			name: "custom auth is explicit and carries its header routing",
			request: McpCreateRequest{
				Name: "demo", Backend: McpBackendExternal,
				Transport: "streamable_http", EndpointURL: &endpoint,
				Auth: McpAuth{
					Type: "custom", Credential: &credential,
					HeaderName: &header, HeaderPrefix: &prefix,
				},
			},
			want: `{"name":"demo","backend":"external",` +
				`"endpoint_url":"https://example.com/mcp","transport":"streamable_http",` +
				`"auth":{"type":"custom","credential":"token",` +
				`"header_name":"X-Token","header_prefix":"Token "}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("json =\n  %s\nwant\n  %s", got, tt.want)
			}
		})
	}
}

func TestFindMcpCatalogEntry(t *testing.T) {
	entries := []McpCatalogEntry{
		{ID: "linear", Name: "Linear"},
		{ID: "notion", Name: "Notion"},
	}
	tests := []struct {
		name      string
		entries   []McpCatalogEntry
		catalogID string
		wantName  string
		wantErr   string
	}{
		{name: "first entry", entries: entries, catalogID: "linear", wantName: "Linear"},
		{name: "later entry", entries: entries, catalogID: "notion", wantName: "Notion"},
		{
			name: "unknown id names the command that lists them", entries: entries,
			catalogID: "nope",
			wantErr:   `no catalog entry "nope" — see 'iai mcps catalog'`,
		},
		{
			name: "an empty catalog is not a panic", entries: nil, catalogID: "linear",
			wantErr: `no catalog entry "linear" — see 'iai mcps catalog'`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := FindMcpCatalogEntry(tt.entries, tt.catalogID)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FindMcpCatalogEntry() error = %v", err)
			}
			if entry.Name != tt.wantName {
				t.Errorf("entry.Name = %q, want %q", entry.Name, tt.wantName)
			}
		})
	}
}

func TestDecodeMcpDetailEnvAndSecretRefs(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantEnv        []McpEnvVar
		wantSecretRefs []string
	}{
		{
			name: "env and secret_refs decode from snake_case",
			body: `{"success":true,"data":{"mcp":{"name":"smoke","backend":"internal",` +
				`"env":[{"name":"API_BASE","value":"https://api.example"},{"name":"DEBUG","value":"1"}],` +
				`"secret_refs":["smoke-token","smoke-basic"]}}}`,
			wantEnv: []McpEnvVar{
				{Name: "API_BASE", Value: "https://api.example"},
				{Name: "DEBUG", Value: "1"},
			},
			wantSecretRefs: []string{"smoke-token", "smoke-basic"},
		},
		{
			name: "a server that omits both leaves them nil",
			body: `{"success":true,"data":{"mcp":{"name":"smoke","backend":"internal"}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := decodeSuccess[McpDetailData]([]byte(tt.body), "describe mcp")
			if err != nil {
				t.Fatalf("decodeSuccess() error = %v", err)
			}
			if diff := cmp.Diff(tt.wantEnv, data.Mcp.Env); diff != "" {
				t.Errorf("env mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSecretRefs, data.Mcp.SecretRefs); diff != "" {
				t.Errorf("secret_refs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
