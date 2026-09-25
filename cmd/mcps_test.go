package cmd

import (
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/spf13/cobra"
)

func TestValidateMcpBackendFlags(t *testing.T) {
	tests := []struct {
		name    string
		backend platform.McpBackend
		flag    string
		wantErr string
	}{
		{name: "internal image", backend: platform.McpBackendInternal, flag: "image-name"},
		{name: "internal endpoint", backend: platform.McpBackendInternal, flag: "endpoint=false"},
		{
			name:    "external endpoint",
			backend: platform.McpBackendExternal,
			flag:    "endpoint=false",
			wantErr: "--endpoint only applies to an internal mcp",
		},
		{
			name:    "external stack clearing",
			backend: platform.McpBackendExternal,
			flag:    "clear-stack-id=true",
			wantErr: "--clear-stack-id only applies to an internal mcp",
		},
		{
			name:    "external image",
			backend: platform.McpBackendExternal,
			flag:    "image-name",
			wantErr: "--image-name only applies to an internal mcp",
		},
		{name: "external auth header", backend: platform.McpBackendExternal, flag: "auth-header"},
		{name: "internal env", backend: platform.McpBackendInternal, flag: "env"},
		{
			name:    "external env",
			backend: platform.McpBackendExternal,
			flag:    "env",
			wantErr: "--env only applies to an internal mcp",
		},
		{name: "internal secret", backend: platform.McpBackendInternal, flag: "secret"},
		{
			name:    "external secret",
			backend: platform.McpBackendExternal,
			flag:    "secret",
			wantErr: "--secret only applies to an internal mcp",
		},
		{
			name:    "external clear-env",
			backend: platform.McpBackendExternal,
			flag:    "clear-env=true",
			wantErr: "--clear-env only applies to an internal mcp",
		},
		{
			name:    "external clear-secret",
			backend: platform.McpBackendExternal,
			flag:    "clear-secret=true",
			wantErr: "--clear-secret only applies to an internal mcp",
		},
		{
			name:    "internal auth header",
			backend: platform.McpBackendInternal,
			flag:    "auth-header",
		},
		{name: "external client id", backend: platform.McpBackendExternal, flag: "client-id"},
		{
			name:    "internal client id",
			backend: platform.McpBackendInternal,
			flag:    "client-id",
			wantErr: "--client-id only applies to an external mcp",
		},
		{
			name:    "internal client secret stdin",
			backend: platform.McpBackendInternal,
			flag:    "client-secret-stdin=true",
			wantErr: "--client-secret-stdin only applies to an external mcp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "mcp"}
			var value string
			for _, name := range []string{
				"image-name", "image-tag", "port", "path", "memory", "cpu",
				"stack-id", "auth-header", "auth-header-prefix", "client-id", "client-secret",
			} {
				cmd.Flags().StringVar(&value, name, "", "")
			}
			cmd.Flags().StringArray("env", nil, "")
			cmd.Flags().StringArray("secret", nil, "")
			cmd.Flags().Bool("clear-env", false, "")
			cmd.Flags().Bool("clear-secret", false, "")
			cmd.Flags().Bool("endpoint", false, "")
			cmd.Flags().Bool("clear-stack-id", false, "")
			cmd.Flags().Bool("client-secret-stdin", false, "")
			name, value, found := strings.Cut(tt.flag, "=")
			if !found {
				value = "x"
			}
			if err := cmd.Flags().Set(name, value); err != nil {
				t.Fatalf("set %s: %v", name, err)
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

func TestValidateMcpCreateAuth(t *testing.T) {
	tests := []struct {
		name     string
		auth     string
		header   string
		prefix   string
		clientID string
		wantErr  string
	}{
		{name: "custom header", auth: "custom", header: "X-Token"},
		{
			name: "custom requires header", auth: "custom",
			wantErr: "--auth-type custom requires --auth-header",
		},
		{
			name: "custom rejects prefix without header", auth: "custom", prefix: "Token ",
			wantErr: "--auth-type custom requires --auth-header",
		},
		{
			name: "bearer rejects header", auth: "bearer", header: "X-Token",
			wantErr: "--auth-header and --auth-header-prefix require --auth-type custom",
		},
		{
			name: "bearer rejects prefix", auth: "bearer", prefix: "Token ",
			wantErr: "--auth-header and --auth-header-prefix require --auth-type custom",
		},
		{name: "client credentials with client id", auth: "client_credentials", clientID: "id"},
		{
			name: "client credentials requires client id", auth: "client_credentials",
			wantErr: "--auth-type client_credentials requires --client-id and --client-secret",
		},
		{
			name:     "bearer rejects client id",
			auth:     "bearer",
			clientID: "id",
			wantErr:  "--client-id and --client-secret require --auth-type oauth or client_credentials",
		},
		{
			name:     "custom rejects client id",
			auth:     "custom",
			header:   "X-Token",
			clientID: "id",
			wantErr:  "--client-id and --client-secret require --auth-type oauth or client_credentials",
		},
		{name: "oauth accepts a customer client id", auth: "oauth", clientID: "id"},
		{name: "oauth still works without one", auth: "oauth"},
		{
			name:     "none rejects client id",
			auth:     "none",
			clientID: "id",
			wantErr:  "--client-id and --client-secret require --auth-type oauth or client_credentials",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "create"}
			cmd.Flags().String("auth-header", "", "")
			cmd.Flags().String("auth-header-prefix", "", "")
			cmd.Flags().String("client-id", "", "")
			for name, value := range map[string]string{
				"auth-header": tt.header, "auth-header-prefix": tt.prefix, "client-id": tt.clientID,
			} {
				if value != "" {
					if err := cmd.Flags().Set(name, value); err != nil {
						t.Fatalf("set --%s: %v", name, err)
					}
				}
			}
			err := validateMcpCreateAuth(cmd, platform.McpAuthType(tt.auth))
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateMcpCreateAuth() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("validateMcpCreateAuth() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMcpUpdateAuth(t *testing.T) {
	tests := []struct {
		name    string
		backend platform.McpBackend
		flags   []string
		wantErr string
	}{
		{
			name:    "internal credential only",
			backend: platform.McpBackendInternal,
			flags:   []string{"credential=x"},
		},
		{
			name:    "internal custom type only",
			backend: platform.McpBackendInternal,
			flags:   []string{"auth-type=custom"},
		},
		{
			name: "external credential requires auth type", backend: platform.McpBackendExternal,
			flags: []string{"credential=x"}, wantErr: "--credential requires --auth-type",
		},
		{
			name:    "external custom requires header",
			backend: platform.McpBackendExternal,
			flags: []string{
				"auth-type=custom",
			},
			wantErr: "--auth-type custom requires --auth-header",
		},
		{
			name:    "external custom rejects empty header",
			backend: platform.McpBackendExternal,
			flags: []string{
				"auth-type=custom",
				"auth-header=",
			},
			wantErr: "--auth-type custom requires --auth-header",
		},
		{
			name: "external bearer rejects header", backend: platform.McpBackendExternal,
			flags:   []string{"auth-type=bearer", "auth-header=X-Token"},
			wantErr: "--auth-header and --auth-header-prefix require --auth-type custom",
		},
		{
			name: "external custom header", backend: platform.McpBackendExternal,
			flags: []string{"auth-type=custom", "auth-header=X-Token"},
		},
		{
			name: "external machine auth cannot be updated", backend: platform.McpBackendExternal,
			flags: []string{"auth-type=client_credentials"},
			wantErr: "client_credentials auth cannot be changed in place; delete the mcp " +
				"and recreate it with --client-id and --client-secret-stdin",
		},
		{
			name: "external oauth rotates its client", backend: platform.McpBackendExternal,
			flags: []string{"auth-type=oauth", "client-id=new-id", "client-secret=new-secret"},
		},
		{
			name: "external oauth rotation needs a secret", backend: platform.McpBackendExternal,
			flags: []string{"auth-type=oauth", "client-id=new-id"},
			wantErr: "rotating an oauth client needs --client-id and --client-secret " +
				"(or --client-secret-stdin) together",
		},
		{
			name: "external oauth rotation needs a client id", backend: platform.McpBackendExternal,
			flags: []string{"auth-type=oauth", "client-secret=new-secret"},
			wantErr: "rotating an oauth client needs --client-id and --client-secret " +
				"(or --client-secret-stdin) together",
		},
		{
			name:    "external oauth rotation takes the secret on stdin",
			backend: platform.McpBackendExternal,
			flags:   []string{"auth-type=oauth", "client-id=new-id", "client-secret-stdin=true"},
		},
		{
			name:    "external client id still needs an auth type",
			backend: platform.McpBackendExternal,
			flags:   []string{"client-id=new-id"},
			wantErr: "--client-id requires --auth-type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "update"}
			for _, name := range []string{
				"auth-type", "credential", "auth-header", "auth-header-prefix",
				"client-id", "client-secret",
			} {
				cmd.Flags().String(name, "", "")
			}
			cmd.Flags().Bool("credential-stdin", false, "")
			cmd.Flags().Bool("client-secret-stdin", false, "")
			for _, flag := range tt.flags {
				name, value, _ := strings.Cut(flag, "=")
				if err := cmd.Flags().Set(name, value); err != nil {
					t.Fatalf("set --%s: %v", name, err)
				}
			}
			err := validateMcpUpdateAuth(cmd, tt.backend)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateMcpUpdateAuth() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("validateMcpUpdateAuth() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestMcpAuthTypeOr(t *testing.T) {
	tests := []struct {
		name       string
		backend    platform.McpBackend
		explicit   string
		credential string
		header     string
		prefix     string
		want       string
	}{
		{
			name:     "explicit wins",
			backend:  platform.McpBackendExternal,
			explicit: "api_key",
			want:     "api_key",
		},
		{
			name:    "custom header",
			backend: platform.McpBackendExternal,
			header:  "X-Token",
			want:    "custom",
		},
		{
			name:    "internal custom header",
			backend: platform.McpBackendInternal,
			header:  "X-Token",
			want:    "custom",
		},
		{
			name:    "prefix also means custom",
			backend: platform.McpBackendExternal,
			prefix:  "Token ",
			want:    "custom",
		},
		{
			name: "external credential defaults to bearer", backend: platform.McpBackendExternal,
			credential: "secret", want: "bearer",
		},
		{
			name:       "internal credential defaults to bearer too",
			backend:    platform.McpBackendInternal,
			credential: "secret",
			want:       "bearer",
		},
		{
			name:    "no auth fields defaults to none",
			backend: platform.McpBackendExternal,
			want:    "none",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mcpAuthTypeOr(tt.backend, tt.explicit, tt.credential, tt.header, tt.prefix, "")
			if string(got) != tt.want {
				t.Errorf("mcpAuthTypeOr() = %q, want %q", got, tt.want)
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
		clientID   string
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
			name:    "an undeclared entry is refused",
			methods: nil,
			cred:    "a-token",
			wantErr: true,
		},
		{
			name:     "a client_credentials entry needs no explicit auth type",
			methods:  []string{"client_credentials"},
			clientID: "an-app-client-id",
			wantAuth: "client_credentials",
		},
		{
			name:     "a client id infers client_credentials and no sign-in",
			methods:  []string{"client_credentials", "oauth"},
			clientID: "an-app-client-id",
			wantAuth: "client_credentials",
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
			got := mcpAuthTypeOr(
				platform.McpBackendExternal, fromCatalog, tt.cred, tt.header, "", tt.clientID,
			)
			if string(got) != tt.wantAuth {
				t.Errorf("resolved auth type = %q, want %q", got, tt.wantAuth)
			}
			if signIn := got == platform.McpAuthOAuth; signIn != tt.wantSignIn {
				t.Errorf("sign-in message = %v, want %v", signIn, tt.wantSignIn)
			}
		})
	}
}

// The backend resolves the issuer, token endpoint and scopes for
// client_credentials from a curated catalog entry, so --client-id on a
// --external-url mcp can only ever reach it as a contradictory payload.
func TestClientIdAndExternalUrlAreMutuallyExclusive(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr bool
	}{
		{
			name:    "client id with an external url",
			flags:   []string{"--external-url", "https://mcp.acme.com/mcp", "--client-id", "abc"},
			wantErr: true,
		},
		{
			name:  "client id with a catalog entry",
			flags: []string{"--catalog-id", "mongodbatlas", "--client-id", "abc"},
		},
		{
			name:  "an external url on its own",
			flags: []string{"--external-url", "https://mcp.acme.com/mcp", "--credential", "t"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mcpCreateCmd.ParseFlags(tt.flags); err != nil {
				t.Fatalf("ParseFlags: %v", err)
			}
			// Parsed flags stick to the shared command, so each case starts clean.
			defer func() {
				for _, name := range []string{"external-url", "catalog-id", "client-id", "credential"} {
					mcpCreateCmd.Flags().Lookup(name).Changed = false
				}
			}()

			err := mcpCreateCmd.ValidateFlagGroups()
			if tt.wantErr && err == nil {
				t.Fatalf("expected the flag combination to be refused")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected refusal: %v", err)
			}
		})
	}
}
