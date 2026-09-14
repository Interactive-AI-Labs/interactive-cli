package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
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
					"ResolveToolArgs(%q) parsed %d entries, want %d",
					tt.inline,
					len(got),
					tt.wantLen,
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

func TestResolveMcpEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		envVars []string
		changed bool
		clear   bool
		want    []platform.McpEnvVar
		wantErr string
	}{
		{name: "untouched leaves the deployed list alone", want: nil},
		{name: "--clear-env clears it", clear: true, want: []platform.McpEnvVar{}},
		{
			name:    "name and value",
			envVars: []string{"ENV=dev"},
			changed: true,
			want:    []platform.McpEnvVar{{Name: "ENV", Value: "dev"}},
		},
		{
			name:    "a value keeps its own equals signs",
			envVars: []string{"DSN=postgres://u:p@h/db?x=1"},
			changed: true,
			want:    []platform.McpEnvVar{{Name: "DSN", Value: "postgres://u:p@h/db?x=1"}},
		},
		{
			name:    "an empty value is a value",
			envVars: []string{"SILENT_MODE="},
			changed: true,
			want:    []platform.McpEnvVar{{Name: "SILENT_MODE", Value: ""}},
		},
		{
			name:    "order is preserved",
			envVars: []string{"B=2", "A=1"},
			changed: true,
			want:    []platform.McpEnvVar{{Name: "B", Value: "2"}, {Name: "A", Value: "1"}},
		},
		{
			name:    "missing equals",
			envVars: []string{"ENV"},
			changed: true,
			wantErr: `invalid --env value "ENV"; expected NAME=VALUE`,
		},
		{
			name:    "empty name",
			envVars: []string{"=dev"},
			changed: true,
			wantErr: `invalid --env value "=dev"; expected NAME=VALUE`,
		},
		{
			name:    "clear with --env is a contradiction",
			envVars: []string{"A=1"},
			changed: true,
			clear:   true,
			wantErr: "--clear-env cannot be combined with --env",
		},
		{
			name:    "--env with no values points at --clear-env",
			changed: true,
			wantErr: "--env requires at least one NAME=VALUE argument; use --clear-env to remove all variables",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveMcpEnvVars(tt.envVars, tt.changed, tt.clear)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveMcpSecretRefs(t *testing.T) {
	tests := []struct {
		name    string
		refs    []string
		changed bool
		clear   bool
		want    []string
		wantErr string
	}{
		{name: "untouched leaves the deployed list alone", want: nil},
		{name: "--clear-secret clears it", clear: true, want: []string{}},
		{
			name:    "names are trimmed and ordered",
			refs:    []string{"platform-dev", " services-dev "},
			changed: true,
			want:    []string{"platform-dev", "services-dev"},
		},
		{
			name:    "a blank name is rejected",
			refs:    []string{" "},
			changed: true,
			wantErr: `invalid --secret value " "; name must not be empty`,
		},
		{
			name:    "clear with --secret is a contradiction",
			refs:    []string{"platform-dev"},
			changed: true,
			clear:   true,
			wantErr: "--clear-secret cannot be combined with --secret",
		},
		{
			name:    "--secret with no values points at --clear-secret",
			changed: true,
			wantErr: "--secret requires at least one secret name; use --clear-secret to remove all secret references",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveMcpSecretRefs(tt.refs, tt.changed, tt.clear)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
