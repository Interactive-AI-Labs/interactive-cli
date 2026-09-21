package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConfirmDeletion(t *testing.T) {
	tests := []struct {
		name       string
		stdin      string
		noTTY      bool
		wantOK     bool
		wantPrompt string
		wantErr    string
	}{
		{
			name:       "yes with newline",
			stdin:      "y\n",
			wantOK:     true,
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:       "yes without newline (EOF mid-line)",
			stdin:      "y",
			wantOK:     true,
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:       "uppercase yes",
			stdin:      "Y\n",
			wantOK:     true,
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:       "yes with surrounding space",
			stdin:      "  y  \n",
			wantOK:     true,
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:       "no",
			stdin:      "n\n",
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:       "empty stdin (bare EOF)",
			stdin:      "",
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:       "anything else declines",
			stdin:      "yes please\n",
			wantPrompt: `This will delete mcp "my-mcp". Continue? [y/N] `,
		},
		{
			name:    "no terminal refuses instead of blocking",
			noTTY:   true,
			wantErr: `cannot ask whether to delete mcp "my-mcp": stdin is not a terminal; re-run with -f to confirm`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in io.Reader = strings.NewReader(tt.stdin)
			if tt.noTTY {
				in = idleFile(t)
			}

			var out bytes.Buffer
			ok, err := confirmWithin(t, in, &out, `mcp "my-mcp"`)

			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("error = %q, want %q", gotErr, tt.wantErr)
			}
			if ok != tt.wantOK {
				t.Errorf("confirmed = %v, want %v", ok, tt.wantOK)
			}
			if got := out.String(); got != tt.wantPrompt {
				t.Errorf("prompt = %q, want %q", got, tt.wantPrompt)
			}
		})
	}
}

// confirmWithin fails fast instead of hanging the suite if the terminal guard regresses.
func confirmWithin(t *testing.T, in io.Reader, out *bytes.Buffer, target string) (bool, error) {
	t.Helper()

	type result struct {
		ok  bool
		err error
	}
	done := make(chan result, 1)
	go func() {
		ok, err := confirmDeletion(in, out, target)
		done <- result{ok, err}
	}()

	select {
	case r := <-done:
		return r.ok, r.err
	case <-time.After(5 * time.Second):
		t.Fatal("confirmDeletion blocked on a reader that will never deliver input")
		return false, nil
	}
}

// idleFile is a real *os.File that is not a terminal and never delivers input,
// like the stdin a CI runner hands the command.
func idleFile(t *testing.T) io.Reader {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	return r
}
