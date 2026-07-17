package inputs

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

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
		wantErr    bool
		errContain string
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
			wantErr: true,
		},
		{
			name: "external rejects secret instead of silently dropping it",
			input: McpInput{
				EndpointURL: "https://mcp.acme.com/mcp",
				SecretRefs:  []string{"some-secret"},
			},
			wantErr:    true,
			errContain: "don't apply to an external mcp",
		},
		{
			name: "external rejects path instead of silently dropping it",
			input: McpInput{
				EndpointURL: "https://mcp.acme.com/mcp",
				Path:        "/other",
			},
			wantErr:    true,
			errContain: "don't apply to an external mcp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildMcpRequestBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Fatalf("err = %q, want containing %q", err, tt.errContain)
				}
				return
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
			wantErr: "expected KEY=VALUE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildMcpRequestBody(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
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
	t.Run("no flags changed produces an empty patch", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(McpInput{}, false, false, false, changedFunc())
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(patch) != 0 {
			t.Fatalf("patch = %v, want empty", patch)
		}
	})

	t.Run("port and path", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(
			McpInput{Port: 9090, Path: "/api/v2"},
			false, false, false,
			changedFunc("port", "path"),
		)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		assertJSONField(t, patch, "port", float64(9090))
		assertJSONField(t, patch, "path", "/api/v2")
	})

	t.Run("partial image only sets the changed sub-fields", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(
			McpInput{ImageTag: "v2"},
			false, false, false,
			changedFunc("image-tag"),
		)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		assertJSONField(t, patch, "image", map[string]any{"tag": "v2"})
	})

	t.Run("env replace requires at least one value", func(t *testing.T) {
		_, err := BuildMcpUpdatePatch(McpInput{}, false, false, false, changedFunc("env"))
		if err == nil || !strings.Contains(err.Error(), "--clear-env") {
			t.Fatalf("err = %v, want mentioning --clear-env", err)
		}
	})

	t.Run("clear-env clears without --env", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(McpInput{}, true, false, false, changedFunc())
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if string(patch["env"]) != "null" {
			t.Fatalf(`patch["env"] = %s, want "null"`, patch["env"])
		}
	})

	t.Run("clear-env combined with --env errors", func(t *testing.T) {
		_, err := BuildMcpUpdatePatch(
			McpInput{EnvVars: []string{"A=1"}},
			true, false, false,
			changedFunc("env"),
		)
		if err == nil ||
			!strings.Contains(err.Error(), "--clear-env cannot be combined with --env") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("auth-type none implicitly clears the credential", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(
			McpInput{AuthType: "none"},
			false, false, false,
			changedFunc("auth-type"),
		)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		assertJSONField(t, patch, "auth", map[string]any{"type": "none", "credential": ""})
	})

	t.Run("credential rotation without changing auth-type", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(
			McpInput{Credential: "new-token"},
			false, false, false,
			changedFunc("credential"),
		)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		assertJSONField(t, patch, "auth", map[string]any{"credential": "new-token"})
	})

	t.Run("headers set", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(
			McpInput{Headers: []string{"X-Team=platform"}},
			false, false, false,
			changedFunc("header"),
		)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		assertJSONField(t, patch, "headers", map[string]any{"X-Team": "platform"})
	})

	t.Run("clear-headers clears", func(t *testing.T) {
		patch, err := BuildMcpUpdatePatch(McpInput{}, false, false, true, changedFunc())
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if string(patch["headers"]) != "null" {
			t.Fatalf(`patch["headers"] = %s, want "null"`, patch["headers"])
		}
	})

	t.Run("clear-headers combined with --header errors", func(t *testing.T) {
		_, err := BuildMcpUpdatePatch(
			McpInput{Headers: []string{"X-Team=platform"}},
			false, false, true,
			changedFunc("header"),
		)
		if err == nil ||
			!strings.Contains(err.Error(), "--clear-headers cannot be combined with --header") {
			t.Fatalf("err = %v", err)
		}
	})
}

func assertJSONField(t *testing.T, patch map[string]json.RawMessage, key string, want any) {
	t.Helper()
	raw, ok := patch[key]
	if !ok {
		t.Fatalf("patch[%q] missing; patch = %v", key, patch)
	}
	var got any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("patch[%q] = %s: %v", key, raw, err)
	}
	// round-trip want through JSON too, so e.g. int vs float64 compare equal
	wantRaw, _ := json.Marshal(want)
	var wantNorm any
	_ = json.Unmarshal(wantRaw, &wantNorm)
	if !reflect.DeepEqual(got, wantNorm) {
		t.Fatalf("patch[%q] = %#v, want %#v", key, got, wantNorm)
	}
}
