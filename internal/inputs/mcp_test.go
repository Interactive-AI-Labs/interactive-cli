package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestParseHeaderFlags(t *testing.T) {
	tests := []struct {
		name    string
		pairs   []string
		want    map[string]string
		wantErr bool
	}{
		{"simple pair", []string{"X-A=1"}, map[string]string{"X-A": "1"}, false},
		{"value with equals", []string{"X-B=two=2"}, map[string]string{"X-B": "two=2"}, false},
		{"multiple", []string{"X-A=1", "X-B=2"}, map[string]string{"X-A": "1", "X-B": "2"}, false},
		{"missing equals", []string{"bad-no-equals"}, nil, true},
		{"empty key", []string{"=value"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeaderFlags(tt.pairs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseHeaderFlags(%v) err=%v wantErr=%v", tt.pairs, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Fatalf("ParseHeaderFlags(%v) = %#v, want %q=%q", tt.pairs, got, k, v)
				}
			}
		})
	}
}

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

func TestCatalogEndpointURL(t *testing.T) {
	entries := []clients.McpCatalogEntry{
		{ID: "kiwi", EndpointURL: "https://mcp.kiwi.com/"},
		{ID: "selfhosted", EndpointURL: ""},
	}
	tests := []struct {
		name      string
		catalogID string
		want      string
		wantErr   bool
	}{
		{"known entry with endpoint", "kiwi", "https://mcp.kiwi.com/", false},
		{"unknown entry", "missing", "", true},
		{"entry without managed endpoint", "selfhosted", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CatalogEndpointURL(entries, tt.catalogID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CatalogEndpointURL(%q) err=%v wantErr=%v", tt.catalogID, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
