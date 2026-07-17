package inputs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveCredential(t *testing.T) {
	tests := []struct {
		name       string
		stdin      string
		credential string
		fromStdin  bool
		want       string
	}{
		{"flag passes through", "ignored", "flag-secret", false, "flag-secret"},
		{"stdin trims trailing LF", "stdin-secret\n", "", true, "stdin-secret"},
		{"stdin trims trailing CRLF", "stdin-secret\r\n", "", true, "stdin-secret"},
		{"stdin keeps inner whitespace", "a b\n", "", true, "a b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCredential(strings.NewReader(tt.stdin), tt.credential, tt.fromStdin)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveToolArgs(t *testing.T) {
	tests := []struct {
		name    string
		inline  string
		wantLen int
		wantKey string
		wantVal string
		wantErr bool
	}{
		{"empty yields empty object", "", 0, "", "", false},
		{"inline object", `{"q":"foo","n":2}`, 2, "q", "foo", false},
		{"non-object array rejected", `[1,2,3]`, 0, "", "", true},
		{"json null rejected", `null`, 0, "", "", true},
		{"invalid json rejected", `{not json}`, 0, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveToolArgs(tt.inline, "")
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveToolArgs(%q) err=%v wantErr=%v", tt.inline, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != tt.wantLen {
				t.Fatalf(
					"ResolveToolArgs(%q) parsed %d entries, want %d: %#v",
					tt.inline,
					len(got),
					tt.wantLen,
					got,
				)
			}
			if tt.wantKey != "" && got[tt.wantKey] != tt.wantVal {
				t.Fatalf("got %#v, want %q=%q", got, tt.wantKey, tt.wantVal)
			}
		})
	}
}

func TestResolveToolArgsFromFile(t *testing.T) {
	dir := t.TempDir()
	objPath := filepath.Join(dir, "args.json")
	if err := os.WriteFile(objPath, []byte(`{"q":"foo","n":2}`), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	arrPath := filepath.Join(dir, "arr.json")
	if err := os.WriteFile(arrPath, []byte(`[1,2,3]`), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	tests := []struct {
		name    string
		file    string
		wantKey string
		wantErr bool
	}{
		{"valid object file", objPath, "q", false},
		{"missing file rejected", filepath.Join(dir, "nope.json"), "", true},
		{"non-object file rejected", arrPath, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveToolArgs("", tt.file)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveToolArgs(file=%q) err=%v wantErr=%v", tt.file, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got[tt.wantKey] != "foo" {
				t.Fatalf("unexpected args: %#v", got)
			}
		})
	}
}

// changedFunc turns a list of flag names into the `changed` predicate BuildMcpUpdatePatch expects.
func changedFunc(names ...string) func(string) bool {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return func(name string) bool { return set[name] }
}

func TestBuildMcpRequestBody(t *testing.T) {
	tests := []struct {
		name       string
		input      McpInput
		wantErr    string
		wantSecret []string
	}{
		{
			name: "internal with secrets",
			input: McpInput{
				Port:       8080,
				ImageName:  "my-mcp",
				ImageTag:   "v1",
				SecretRefs: []string{"db-creds", "api-key"},
			},
			wantSecret: []string{"db-creds", "api-key"},
		},
		{
			name: "internal rejects empty secret name",
			input: McpInput{
				Port:       8080,
				ImageName:  "my-mcp",
				ImageTag:   "v1",
				SecretRefs: []string{"  "},
			},
			wantErr: `invalid --secret value "  "; name must not be empty`,
		},
		{
			name: "external rejects secret instead of silently dropping it",
			input: McpInput{
				EndpointURL: "https://mcp.acme.com/mcp",
				SecretRefs:  []string{"some-secret"},
			},
			wantErr: "--env, --secret, and --path don't apply to an external mcp — the path is part of --external-url",
		},
		{
			name: "external rejects path instead of silently dropping it",
			input: McpInput{
				EndpointURL: "https://mcp.acme.com/mcp",
				Path:        "/other",
			},
			wantErr: "--env, --secret, and --path don't apply to an external mcp — the path is part of --external-url",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildMcpRequestBody(tt.input)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if len(body.SecretRefs) != len(tt.wantSecret) {
				t.Fatalf("SecretRefs = %+v, want %v", body.SecretRefs, tt.wantSecret)
			}
			for i, want := range tt.wantSecret {
				if body.SecretRefs[i].SecretName != want {
					t.Errorf("SecretRefs[%d] = %q, want %q", i, body.SecretRefs[i].SecretName, want)
				}
			}
		})
	}
}

func TestBuildMcpRequestBodyAuth(t *testing.T) {
	tests := []struct {
		name         string
		input        McpInput
		wantAuthType string
		wantHeaders  map[string]string
		wantErr      string
	}{
		{
			name:         "credential infers bearer",
			input:        McpInput{EndpointURL: "https://x.io/mcp", Credential: "tok"},
			wantAuthType: "bearer",
		},
		{
			name:         "no credential infers none",
			input:        McpInput{EndpointURL: "https://x.io/mcp"},
			wantAuthType: "none",
		},
		{
			name: "explicit auth-type wins over inference",
			input: McpInput{
				EndpointURL: "https://x.io/mcp",
				Credential:  "k",
				AuthType:    "api_key",
			},
			wantAuthType: "api_key",
		},
		{
			name: "auth-header without auth-type infers custom",
			input: McpInput{
				EndpointURL: "https://x.io/mcp",
				Credential:  "k",
				AuthHeader:  "X-Api-Token",
			},
			wantAuthType: "custom",
		},
		{
			name: "headers parsed into a map",
			input: McpInput{
				EndpointURL: "https://x.io/mcp",
				Headers:     []string{"X-Org=acme", "X-Trace=1"},
			},
			wantAuthType: "none",
			wantHeaders:  map[string]string{"X-Org": "acme", "X-Trace": "1"},
		},
		{
			name:    "malformed header rejected",
			input:   McpInput{EndpointURL: "https://x.io/mcp", Headers: []string{"bogus"}},
			wantErr: `invalid --header "bogus": expected KEY=VALUE`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildMcpRequestBody(tt.input)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if body.Auth.Type != tt.wantAuthType {
				t.Errorf("Auth.Type = %q, want %q", body.Auth.Type, tt.wantAuthType)
			}
			if len(body.Headers) != len(tt.wantHeaders) {
				t.Fatalf("Headers = %v, want %v", body.Headers, tt.wantHeaders)
			}
			for k, want := range tt.wantHeaders {
				if body.Headers[k] != want {
					t.Errorf("Headers[%q] = %q, want %q", k, body.Headers[k], want)
				}
			}
		})
	}
}

func TestBuildMcpRequestBodyPath(t *testing.T) {
	t.Run("internal path passes through as given, not guessed", func(t *testing.T) {
		body, err := BuildMcpRequestBody(McpInput{
			Port: 8080, ImageName: "my-mcp", ImageTag: "v1", Path: "/api/v2/mcp",
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if body.Path != "/api/v2/mcp" {
			t.Errorf("Path = %q, want %q", body.Path, "/api/v2/mcp")
		}
	})
	t.Run("internal with no --path leaves it empty for the server to default", func(t *testing.T) {
		body, err := BuildMcpRequestBody(McpInput{Port: 8080, ImageName: "my-mcp", ImageTag: "v1"})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if body.Path != "" {
			t.Errorf("Path = %q, want empty (CLI must not guess)", body.Path)
		}
	})
}

func TestBuildMcpUpdatePatch(t *testing.T) {
	tests := []struct {
		name         string
		input        McpInput
		clearEnv     bool
		clearSecret  bool
		clearHeaders bool
		changed      []string
		want         map[string]any
		wantErr      string
	}{
		{
			name: "no flags changed produces an empty patch",
			want: map[string]any{},
		},
		{
			name:    "port and path",
			input:   McpInput{Port: 9090, Path: "/api/v2"},
			changed: []string{"port", "path"},
			want:    map[string]any{"port": float64(9090), "path": "/api/v2"},
		},
		{
			name:    "partial image only sets the changed sub-fields",
			input:   McpInput{ImageTag: "v2"},
			changed: []string{"image-tag"},
			want:    map[string]any{"image": map[string]any{"tag": "v2"}},
		},
		{
			name:    "env replace requires at least one value",
			changed: []string{"env"},
			wantErr: "--env requires at least one NAME=VALUE argument; use --clear-env to remove all variables",
		},
		{
			name:     "clear-env clears without --env",
			clearEnv: true,
			want:     map[string]any{"env": nil},
		},
		{
			name:     "clear-env combined with --env errors",
			input:    McpInput{EnvVars: []string{"A=1"}},
			clearEnv: true,
			changed:  []string{"env"},
			wantErr:  "--clear-env cannot be combined with --env",
		},
		{
			name:    "auth-type none implicitly clears the credential",
			input:   McpInput{AuthType: "none"},
			changed: []string{"auth-type"},
			want:    map[string]any{"auth": map[string]any{"type": "none", "credential": ""}},
		},
		{
			name:    "credential rotation without changing auth-type",
			input:   McpInput{Credential: "new-token"},
			changed: []string{"credential"},
			want:    map[string]any{"auth": map[string]any{"credential": "new-token"}},
		},
		{
			name:    "headers set",
			input:   McpInput{Headers: []string{"X-Team=platform"}},
			changed: []string{"header"},
			want:    map[string]any{"headers": map[string]any{"X-Team": "platform"}},
		},
		{
			name:         "clear-headers clears",
			clearHeaders: true,
			want:         map[string]any{"headers": nil},
		},
		{
			name:         "clear-headers combined with --header errors",
			input:        McpInput{Headers: []string{"X-Team=platform"}},
			clearHeaders: true,
			changed:      []string{"header"},
			wantErr:      "--clear-headers cannot be combined with --header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := BuildMcpUpdatePatch(
				tt.input, tt.clearEnv, tt.clearSecret, tt.clearHeaders, changedFunc(tt.changed...),
			)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			got := make(map[string]any, len(patch))
			for k, raw := range patch {
				var v any
				if err := json.Unmarshal(raw, &v); err != nil {
					t.Fatalf("decode patch[%q]: %v", k, err)
				}
				got[k] = v
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("patch mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
