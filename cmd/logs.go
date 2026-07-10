package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const logFollowMaxConnectionTime = 10 * time.Minute

func logFollowContext(ctx context.Context, follow bool) (context.Context, func()) {
	if !follow {
		return ctx, func() {}
	}

	ctx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	ctx, cancelTimeout := context.WithTimeout(ctx, logFollowMaxConnectionTime)
	return ctx, func() {
		cancelTimeout()
		stopSignals()
	}
}

func finishLogStream(errOut io.Writer, follow bool, ctx context.Context, err error) error {
	if !follow {
		return err
	}

	switch ctx.Err() {
	case context.Canceled:
		return nil
	case context.DeadlineExceeded:
		fmt.Fprintf(
			errOut,
			"Maximum connection time is %g minutes, closing connection.\n",
			logFollowMaxConnectionTime.Minutes(),
		)
		return nil
	default:
		return err
	}
}
