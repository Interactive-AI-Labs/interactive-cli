package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/spf13/cobra"
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

func TestRequireWholeWorkload(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr bool
	}{
		{
			name:  "no workload flag is none of its business",
			flags: []string{"description"},
		},
		{
			// The bug this guards: a tag bump used to send image ":v2" with default
			// port, path, memory and cpu, and the update replaces the whole workload.
			name:    "image-tag alone would reset the rest",
			flags:   []string{"image-tag"},
			wantErr: true,
		},
		{
			name:    "memory alone carries no image at all",
			flags:   []string{"memory"},
			wantErr: true,
		},
		{
			name:  "image name and tag together are enough to name the image",
			flags: []string{"image-name", "image-tag"},
		},
		{
			name:  "the whole workload passes",
			flags: []string{"image-name", "image-tag", "port", "path", "memory", "cpu"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "update"}
			var s string
			var n int
			cmd.Flags().StringVar(&s, "description", "", "")
			cmd.Flags().StringVar(&s, "image-name", "", "")
			cmd.Flags().StringVar(&s, "image-tag", "", "")
			cmd.Flags().IntVar(&n, "port", 0, "")
			cmd.Flags().StringVar(&s, "path", "", "")
			cmd.Flags().StringVar(&s, "memory", "", "")
			cmd.Flags().StringVar(&s, "cpu", "", "")
			for _, f := range tt.flags {
				if err := cmd.Flags().Set(f, flagValueFor(f)); err != nil {
					t.Fatalf("set %s: %v", f, err)
				}
			}
			err := requireWholeWorkload(cmd)
			if tt.wantErr && err == nil {
				t.Errorf("requireWholeWorkload() = nil, want an error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("requireWholeWorkload() = %v, want nil", err)
			}
		})
	}
}

func flagValueFor(name string) string {
	if name == "port" {
		return "3000"
	}
	return "x"
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
			// The bug this guards: an undeclared entry used to be refused outright,
			// which made 94 of the 188 dev entries uncreatable from the CLI while
			// the API accepted the same create.
			name:    "undeclared methods ask for the type",
			methods: nil,
			wantErr: true,
		},
		{
			name:     "undeclared methods are fine once given",
			methods:  nil,
			explicit: "oauth",
			want:     "oauth",
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

// TestCatalogCreateAuthResolution covers the composition the create command does:
// resolve the entry's auth type, then fall back to the flag-level default. Two
// bugs lived in the seam rather than in either function.
func TestCatalogCreateAuthResolution(t *testing.T) {
	tests := []struct {
		name       string
		methods    []string
		explicit   string
		cred       string
		wantAuth   string
		wantSignIn bool
		wantErr    bool
	}{
		{
			// A credential used to skip the catalog lookup entirely, so the
			// entry's only method never applied and the create was refused as
			// "does not support auth type 'bearer'".
			name:     "api_key entry with a credential resolves to api_key",
			methods:  []string{"api_key"},
			cred:     "a-key",
			wantAuth: "api_key",
		},
		{
			// And the message used to promise a sign-in for any catalog create
			// without a credential, including an entry that has no sign-in.
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
			name:    "an undeclared entry asks for the type",
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
			got := mcpAuthTypeOr(platform.McpBackendExternal, fromCatalog, tt.cred)
			if got != tt.wantAuth {
				t.Errorf("resolved auth type = %q, want %q", got, tt.wantAuth)
			}
			if signIn := got == "oauth"; signIn != tt.wantSignIn {
				t.Errorf("sign-in message = %v, want %v", signIn, tt.wantSignIn)
			}
		})
	}
}
