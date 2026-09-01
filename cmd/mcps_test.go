package cmd

import (
	"bytes"
	"strings"
	"testing"

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
