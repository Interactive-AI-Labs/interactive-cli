package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

const watchInterval = 2 * time.Second

// runWatch re-runs render every watchInterval until Ctrl-C, redrawing in place.
func runWatch(cmd *cobra.Command, render func(context.Context, io.Writer) error) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	out := cmd.OutOrStdout()
	tty := output.IsTerminal(out)
	for {
		var frame bytes.Buffer
		if err := render(ctx, &frame); err != nil {
			if ctx.Err() != nil {
				return nil // Ctrl-C during an in-flight request
			}
			return err
		}
		if tty {
			io.WriteString(out, "\033[H"+output.Redraw(frame.String()))
		} else {
			out.Write(frame.Bytes())
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
