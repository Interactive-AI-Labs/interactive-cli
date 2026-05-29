package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/spf13/cobra"
)

// Confirmation must honor input without a trailing newline and treat empty input
// as a decline, not an error.
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
			ok, err := confirmDeletion(strings.NewReader(tt.stdin), &out, "some-id")
			if err != nil {
				t.Fatalf("stdin=%q unexpected error: %v", tt.stdin, err)
			}
			if ok != tt.wantOK {
				t.Fatalf("stdin=%q ok=%v wantOK=%v", tt.stdin, ok, tt.wantOK)
			}
			if !strings.Contains(out.String(), "some-id") {
				t.Fatalf("prompt did not mention the connection id: %q", out.String())
			}
		})
	}
}

// Each command must reject an empty or whitespace-only positional arg before
// doing any network work.
func TestIntegrationsEmptyArgGuards(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
		want string
	}{
		{"get", makeIntegrationsGetCmd(), []string{"  "}, "connection id is required"},
		{"verify", makeIntegrationsVerifyCmd(), []string{"  "}, "connection id is required"},
		{"delete", makeIntegrationsDeleteCmd(), []string{"  "}, "connection id is required"},
		{
			"create-custom",
			makeIntegrationsCreateCustomCmd(),
			[]string{"  "},
			"connection name is required",
		},
		{
			"create-from-catalog",
			makeIntegrationsCreateFromCatalogCmd(),
			[]string{"  "},
			"connection name is required",
		},
		{
			"tools run empty id",
			makeIntegrationsToolsRunCmd(),
			[]string{"  ", "search"},
			"connection id is required",
		},
		{
			"tools run empty tool",
			makeIntegrationsToolsRunCmd(),
			[]string{"c1", "  "},
			"tool name is required",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.RunE(tc.cmd, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
		})
	}
}

func TestCatalogEndpointURL(t *testing.T) {
	entries := []clients.McpCatalogEntry{
		{ID: "kiwi", EndpointURL: "https://mcp.kiwi.com/"},
		{ID: "selfhosted", EndpointURL: ""},
	}

	got, err := catalogEndpointURL(entries, "kiwi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://mcp.kiwi.com/" {
		t.Fatalf("got %q, want kiwi endpoint", got)
	}

	if _, err := catalogEndpointURL(entries, "missing"); err == nil {
		t.Fatal("expected error for unknown catalog id")
	}
	if _, err := catalogEndpointURL(entries, "selfhosted"); err == nil {
		t.Fatal("expected error for entry without a managed endpoint")
	}
}

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
				t.Fatalf(
					"validateMcpAuth(%q,%q) err=%v wantErr=%v",
					tt.authType,
					tt.credential,
					err,
					tt.wantErr,
				)
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
	if _, err := parseHeaderFlags([]string{"=value"}); err == nil {
		t.Fatal("expected error for header with empty key")
	}
}

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

	// JSON null rejected (would otherwise unmarshal to a nil map)
	if _, err := resolveToolArgs(`null`, ""); err == nil {
		t.Fatal("expected error for JSON null")
	}

	// invalid JSON rejected
	if _, err := resolveToolArgs(`{not json}`, ""); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestResolveCredential(t *testing.T) {
	// flag value passes through unchanged when not reading stdin
	got, err := resolveCredential(strings.NewReader("ignored"), "flag-secret", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "flag-secret" {
		t.Fatalf("got %q, want flag-secret", got)
	}

	// stdin value is read and a single trailing newline trimmed
	got, err = resolveCredential(strings.NewReader("stdin-secret\n"), "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "stdin-secret" {
		t.Fatalf("got %q, want stdin-secret", got)
	}

	// CRLF trailing newline also trimmed
	got, err = resolveCredential(strings.NewReader("stdin-secret\r\n"), "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "stdin-secret" {
		t.Fatalf("got %q, want stdin-secret", got)
	}
}

func TestResolveToolArgsFromFile(t *testing.T) {
	dir := t.TempDir()

	// valid JSON object file
	objPath := filepath.Join(dir, "args.json")
	if err := os.WriteFile(objPath, []byte(`{"q":"foo","n":2}`), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	got, err := resolveToolArgs("", objPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["q"] != "foo" {
		t.Fatalf("unexpected args: %#v", got)
	}

	// missing file rejected
	if _, err := resolveToolArgs("", filepath.Join(dir, "does-not-exist.json")); err == nil {
		t.Fatal("expected error for missing file")
	}

	// non-object JSON file rejected
	arrPath := filepath.Join(dir, "arr.json")
	if err := os.WriteFile(arrPath, []byte(`[1,2,3]`), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := resolveToolArgs("", arrPath); err == nil {
		t.Fatal("expected error for non-object JSON file")
	}
}
