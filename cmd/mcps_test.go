package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestConfirmDeletion(t *testing.T) {
	tests := []struct {
		name   string
		stdin  string
		wantOK bool
	}{
		{"yes with newline", "y\n", true},
		{"yes without newline (EOF mid-line)", "y", true},
		{"uppercase yes", "Y\n", true},
		{"yes with surrounding space", "  y  \n", true},
		{"no", "n\n", false},
		{"empty stdin (bare EOF)", "", false},
		{"anything else declines", "yes please\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			ok, err := confirmDeletion(strings.NewReader(tt.stdin), &out, `mcp "my-mcp"`)
			if err != nil {
				t.Fatalf("stdin=%q unexpected error: %v", tt.stdin, err)
			}
			if ok != tt.wantOK {
				t.Fatalf("stdin=%q ok=%v wantOK=%v", tt.stdin, ok, tt.wantOK)
			}
			if !strings.Contains(out.String(), "my-mcp") {
				t.Fatalf("prompt did not mention the mcp name: %q", out.String())
			}
		})
	}
}

func TestValidateMcpBackendFlags(t *testing.T) {
	tests := []struct {
		name    string
		backend platform.McpBackend
		flag    string
		wantErr string
	}{
		{name: "internal image", backend: platform.McpBackendInternal, flag: "image-name"},
		{
			name:    "external image",
			backend: platform.McpBackendExternal,
			flag:    "image-name",
			wantErr: "--image-name only applies to an internal mcp",
		},
		{name: "external auth header", backend: platform.McpBackendExternal, flag: "auth-header"},
		{
			name:    "internal auth header",
			backend: platform.McpBackendInternal,
			flag:    "auth-header",
			wantErr: "--auth-header only applies to an external mcp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "mcp"}
			var value string
			for _, name := range []string{
				"image-name", "image-tag", "port", "path", "memory", "cpu",
				"stack-id", "auth-header", "auth-header-prefix",
			} {
				cmd.Flags().StringVar(&value, name, "", "")
			}
			if err := cmd.Flags().Set(tt.flag, "x"); err != nil {
				t.Fatalf("set %s: %v", tt.flag, err)
			}
			err := validateMcpBackendFlags(cmd, tt.backend)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateMcpBackendFlags() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("validateMcpBackendFlags() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMcpUpdateFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr string
	}{
		{name: "no fields", wantErr: "no fields to update; pass at least one flag"},
		{name: "description only", flags: []string{"description"}},
		{
			name:    "image tag requires image name",
			flags:   []string{"image-tag"},
			wantErr: "--image-name and --image-tag must be passed together",
		},
		{
			name:    "image name requires image tag",
			flags:   []string{"image-name"},
			wantErr: "--image-name and --image-tag must be passed together",
		},
		{name: "image name and tag", flags: []string{"image-name", "image-tag"}},
		{name: "memory is independent", flags: []string{"memory"}},
		{
			name:    "empty memory is rejected",
			flags:   []string{"memory="},
			wantErr: "--memory must not be empty",
		},
		{
			name:    "credential requires auth type",
			flags:   []string{"credential"},
			wantErr: "--credential requires --auth-type",
		},
		{name: "credential with auth type", flags: []string{"credential", "auth-type"}},
		{
			name:    "empty stack id is rejected",
			flags:   []string{"stack-id="},
			wantErr: "--stack-id must not be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "update"}
			var value string
			for _, name := range []string{
				"description", "image-name", "image-tag", "memory", "credential",
				"auth-type", "auth-header", "auth-header-prefix", "stack-id",
			} {
				cmd.Flags().StringVar(&value, name, "", "")
			}
			cmd.Flags().Bool("credential-stdin", false, "")
			for _, flag := range tt.flags {
				name, value, found := strings.Cut(flag, "=")
				if !found {
					value = "x"
				}
				if err := cmd.Flags().Set(name, value); err != nil {
					t.Fatalf("set %s: %v", name, err)
				}
			}
			err := validateMcpUpdateFlags(cmd)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateMcpUpdateFlags() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("validateMcpUpdateFlags() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestMcpUpdateRequest(t *testing.T) {
	tests := []struct {
		name      string
		flags     map[string]string
		wantPatch map[string]any
	}{
		{
			name:      "explicit zero port reaches the API",
			flags:     map[string]string{"port": "0"},
			wantPatch: map[string]any{"workload": map[string]any{"port": float64(0)}},
		},
		{
			name:      "memory update stays partial",
			flags:     map[string]string{"memory": "1G"},
			wantPatch: map[string]any{"workload": map[string]any{"memory": "1G"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPatch map[string]any
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.Method == http.MethodGet && r.URL.Path == "/api/v1/session/organizations":
						fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
					case r.Method == http.MethodGet &&
						r.URL.Path == "/api/v1/session/organizations/org-1/projects":
						fmt.Fprint(w, `{"projects":[{"id":"project-1","name":"demo"}]}`)
					case r.Method == http.MethodGet &&
						r.URL.Path == "/api/platform/v1/organizations/org-1/projects/project-1/mcps/tools":
						fmt.Fprint(
							w,
							`{"success":true,"data":{"mcp":{"name":"tools","backend":"internal"}}}`,
						)
					case r.Method == http.MethodPatch &&
						r.URL.Path == "/api/platform/v1/organizations/org-1/projects/project-1/mcps/tools":
						if err := json.NewDecoder(r.Body).Decode(&gotPatch); err != nil {
							t.Errorf("decode request: %v", err)
						}
						fmt.Fprint(
							w,
							`{"success":true,"data":{"mcp":{"name":"tools","backend":"internal"}}}`,
						)
					default:
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}),
			)
			t.Cleanup(server.Close)

			resetFlags := func() {
				mcpUpdateCmd.Flags().VisitAll(func(flag *pflag.Flag) {
					if err := flag.Value.Set(flag.DefValue); err != nil {
						t.Errorf("reset --%s: %v", flag.Name, err)
					}
					flag.Changed = false
				})
			}
			origHostname, origToken, origAPIKey := hostname, token, apiKey
			origOrg, origProject := mcpOrganization, mcpProject
			t.Cleanup(func() {
				hostname, token, apiKey = origHostname, origToken, origAPIKey
				mcpOrganization, mcpProject = origOrg, origProject
				resetFlags()
				mcpUpdateCmd.SetOut(nil)
			})
			hostname, token, apiKey = server.URL, "test-token", ""
			mcpOrganization, mcpProject = "acme", "demo"
			resetFlags()
			for name, value := range tt.flags {
				if err := mcpUpdateCmd.Flags().Set(name, value); err != nil {
					t.Fatalf("set --%s: %v", name, err)
				}
			}
			var out bytes.Buffer
			mcpUpdateCmd.SetOut(&out)
			mcpUpdateCmd.SetContext(context.Background())
			if err := mcpUpdateCmd.RunE(mcpUpdateCmd, []string{"tools"}); err != nil {
				t.Fatalf("mcps update: %v", err)
			}
			if diff := cmp.Diff(tt.wantPatch, gotPatch); diff != "" {
				t.Errorf("patch mismatch (-want +got):\n%s", diff)
			}
			if got, want := out.String(), "Updated tools — internal\n"; got != want {
				t.Errorf("output = %q, want %q", got, want)
			}
		})
	}
}

func TestCatalogAuthType(t *testing.T) {
	tests := []struct {
		name     string
		methods  []string
		explicit string
		want     string
		wantErr  bool
	}{
		{
			name:     "explicit wins over the entry",
			methods:  []string{"bearer"},
			explicit: "oauth",
			want:     "oauth",
		},
		{
			name:    "a single declared method is inferred",
			methods: []string{"bearer"},
			want:    "bearer",
		},
		{
			name:    "several methods leave the choice to the caller",
			methods: []string{"bearer", "api_key"},
			want:    "",
		},
		{
			name:    "undeclared methods are refused, not asked about",
			methods: nil,
			wantErr: true,
		},
		{
			name:     "an explicit type does not rescue an undeclared entry",
			methods:  nil,
			explicit: "oauth",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &platform.McpCatalogEntry{ID: "example", AuthMethods: tt.methods}
			got, err := catalogAuthType(entry, tt.explicit)
			if (err != nil) != tt.wantErr {
				t.Fatalf("catalogAuthType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("catalogAuthType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCatalogCreateAuthResolution(t *testing.T) {
	tests := []struct {
		name       string
		methods    []string
		explicit   string
		cred       string
		header     string
		wantAuth   string
		wantSignIn bool
		wantErr    bool
	}{
		{
			name:     "api_key entry with a credential resolves to api_key",
			methods:  []string{"api_key"},
			cred:     "a-key",
			wantAuth: "api_key",
		},
		{
			name:     "none entry needs no credential and no sign-in",
			methods:  []string{"none"},
			wantAuth: "none",
		},
		{
			name:       "oauth entry is the only one told to sign in",
			methods:    []string{"oauth"},
			wantAuth:   "oauth",
			wantSignIn: true,
		},
		{
			name:     "several methods with a credential fall back to bearer",
			methods:  []string{"bearer", "api_key"},
			cred:     "a-token",
			wantAuth: "bearer",
		},
		{
			name:     "several methods with no credential fall back to none",
			methods:  []string{"bearer", "none"},
			wantAuth: "none",
		},
		{
			name:     "an explicit type still wins",
			methods:  []string{"api_key"},
			explicit: "none",
			wantAuth: "none",
		},
		{
			name:     "a custom header infers bearer for an external mcp",
			methods:  []string{"bearer", "api_key"},
			header:   "X-API-Key",
			wantAuth: "bearer",
		},
		{
			name:    "an undeclared entry is refused",
			methods: nil,
			cred:    "a-token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &platform.McpCatalogEntry{ID: "example", AuthMethods: tt.methods}
			fromCatalog, err := catalogAuthType(entry, tt.explicit)
			if (err != nil) != tt.wantErr {
				t.Fatalf("catalogAuthType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			got := mcpAuthTypeOr(platform.McpBackendExternal, fromCatalog, tt.cred, tt.header, "")
			if got != tt.wantAuth {
				t.Errorf("resolved auth type = %q, want %q", got, tt.wantAuth)
			}
			if signIn := got == "oauth"; signIn != tt.wantSignIn {
				t.Errorf("sign-in message = %v, want %v", signIn, tt.wantSignIn)
			}
		})
	}
}
