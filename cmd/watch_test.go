package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
)

// cmdWithCtx returns a command with the given context and a non-TTY buffer output.
func cmdWithCtx(ctx context.Context, out *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.SetOut(out)
	return cmd
}

func TestRunWatchStopsOnCancelAndSkipsClearForNonTTY(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	cmd := cmdWithCtx(ctx, &out)

	calls := 0
	err := runWatch(cmd, func(_ context.Context, _ io.Writer) error {
		calls++
		cancel() // simulate Ctrl-C after the first render
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil on cancel, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected render called once, got %d", calls)
	}
	if bytes.Contains(out.Bytes(), []byte("\033[")) {
		t.Fatalf("no ANSI escapes should be written to a non-TTY writer")
	}
}

func TestRedrawOverwritesInPlaceWithoutFullClear(t *testing.T) {
	got := redraw("NAME\nfoo\n")
	if bytes.Contains([]byte(got), []byte("\033[2J")) {
		t.Fatalf("redraw must not full-clear (that causes the blink): %q", got)
	}
	want := "NAME\033[K\nfoo\033[K\n\033[J"
	if got != want {
		t.Fatalf("redraw = %q, want %q", got, want)
	}
}

func TestRunWatchCleanExitOnCancelDuringRender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	cmd := cmdWithCtx(ctx, &out)

	err := runWatch(cmd, func(c context.Context, _ io.Writer) error {
		cancel()       // Ctrl-C mid-request
		return c.Err() // HTTP client surfaces context.Canceled
	})
	if err != nil {
		t.Fatalf("expected clean exit when cancelled mid-render, got %v", err)
	}
}

func TestRunWatchAbortsOnRenderError(t *testing.T) {
	var out bytes.Buffer
	cmd := cmdWithCtx(context.Background(), &out)

	want := errors.New("boom")
	err := runWatch(cmd, func(context.Context, io.Writer) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("expected render error to propagate, got %v", err)
	}
}
