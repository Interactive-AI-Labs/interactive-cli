package conversation

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestFollowError(t *testing.T) {
	again := "iai agents events list hotel s1 --min-offset 15 --follow"
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "finished"},
		{name: "stopped by the user", err: context.Canceled},
		{
			name: "stopped while wrapped",
			err:  fmt.Errorf("event request failed: %w", context.Canceled),
		},
		{
			name: "stream ended says how to continue",
			err:  deployment.ErrEventStreamEnded,
			want: "event stream ended; follow it again with: " + again,
		},
		{
			name: "other failures unchanged",
			err:  errors.New("agent returned 404: Session not found (id='s1')"),
			want: "agent returned 404: Session not found (id='s1')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ""
			if err := followError(tt.err, again); err != nil {
				got = err.Error()
			}
			if got != tt.want {
				t.Errorf("error = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFollowEventsCommand(t *testing.T) {
	target := Target{AgentName: "hotel", SessionID: "s1"}
	tests := []struct {
		name       string
		opts       deployment.AgentEventListOptions
		outputFlag string
		want       string
	}{
		{
			name: "offset only",
			opts: deployment.AgentEventListOptions{MinOffset: 15},
			want: "iai agents events list hotel s1 --min-offset 15 --follow",
		},
		{
			name: "filters and output kept",
			opts: deployment.AgentEventListOptions{
				MinOffset: 15,
				Kinds:     "message,status",
				Source:    "customer",
				TraceID:   "t1",
			},
			outputFlag: "--json",
			want: "iai agents events list hotel s1 --min-offset 15 --kinds message,status " +
				"--source customer --trace-id t1 --follow --json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := followEventsCommand(target, tt.opts, tt.outputFlag)
			if got != tt.want {
				t.Errorf("command = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFollowEventCommand(t *testing.T) {
	target := Target{AgentName: "hotel", SessionID: "s1"}
	tests := []struct {
		name       string
		outputFlag string
		want       string
	}{
		{name: "text", want: "iai agents events get hotel s1 e1 --follow"},
		{
			name:       "output kept",
			outputFlag: "--yaml",
			want:       "iai agents events get hotel s1 e1 --follow --yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := followEventCommand(target, "e1", tt.outputFlag)
			if got != tt.want {
				t.Errorf("command = %q, want %q", got, tt.want)
			}
		})
	}
}
