package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestFinishLogStream(t *testing.T) {
	readErr := errors.New("stream closed")

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancelDeadline()
	<-deadlineCtx.Done()

	tests := []struct {
		name    string
		follow  bool
		ctx     context.Context
		err     error
		wantOut string
		wantErr string
	}{
		{
			name:    "non-follow returns read error",
			ctx:     context.Background(),
			err:     readErr,
			wantErr: "stream closed",
		},
		{
			name:   "follow ignores user cancel",
			follow: true,
			ctx:    canceledCtx,
			err:    readErr,
		},
		{
			name:    "follow prints max connection message on deadline",
			follow:  true,
			ctx:     deadlineCtx,
			err:     readErr,
			wantOut: "Maximum connection time is 10 minutes, closing connection.\n",
		},
		{
			name:    "follow returns other read error",
			follow:  true,
			ctx:     context.Background(),
			err:     readErr,
			wantErr: "stream closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := finishLogStream(&out, tt.follow, tt.ctx, tt.err)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if diff := cmp.Diff(tt.wantErr, gotErr); diff != "" {
				t.Fatalf("error mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantOut, out.String()); diff != "" {
				t.Fatalf("output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
