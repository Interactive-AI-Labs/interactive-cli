package cmd

import (
	"bytes"
	"strings"
	"testing"
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
			ok, err := confirmDeletion(strings.NewReader(tt.stdin), &out, "my-mcp")
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
