package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

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
				t.Fatalf("validateMcpAuth(%q,%q) err=%v wantErr=%v", tt.authType, tt.credential, err, tt.wantErr)
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

	// invalid JSON rejected
	if _, err := resolveToolArgs(`{not json}`, ""); err == nil {
		t.Fatal("expected error for invalid JSON")
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
