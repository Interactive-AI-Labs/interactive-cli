// Package conversation drives a conversation session on an agent from the CLI:
// following its events as the agent records them, and the turn-based chat.
package conversation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

// Target names a conversation session on an agent.
type Target struct {
	OrgID, ProjectID, AgentName, SessionID string
}

// errDone stops a follow from a callback without reporting a failure.
var errDone = errors.New("done following")

// FollowEvents shows a session's new events with print until stopped or the stream
// ends, which says how to follow again from where it left off; outputFlag is the
// --json or --yaml flag in use, if any.
func FollowEvents(
	ctx context.Context,
	api *deployment.DeploymentClient,
	target Target,
	opts deployment.AgentEventListOptions,
	outputFlag string,
	print func(ev deployment.AgentEvent, streaming bool) error,
) error {
	next, err := follow(ctx, api, target, opts, print)
	opts.MinOffset = next
	return followError(err, followEventsCommand(target, opts, outputFlag))
}

// FollowEvent shows one event's updates with print until it is complete.
func FollowEvent(
	ctx context.Context,
	api *deployment.DeploymentClient,
	target Target,
	eventID, outputFlag string,
	print func(ev deployment.AgentEvent, streaming bool) error,
) error {
	final, err := api.FollowAgentEvent(
		ctx, target.OrgID, target.ProjectID, target.AgentName, target.SessionID, eventID,
		func(ev deployment.AgentEvent) error {
			if ev.Streaming() {
				return print(ev, true)
			}
			return nil
		},
	)
	if err != nil {
		return followError(err, followEventCommand(target, eventID, outputFlag))
	}
	return print(*final, false)
}

// follow delivers each new event, and every update of a message the agent is still
// writing with streaming set, until ctx ends, the callback returns an error, or the
// stream ends. It returns the offset to resume from, which stays on a message until
// it is complete so a resumed follow shows the rest of it.
func follow(
	ctx context.Context,
	api *deployment.DeploymentClient,
	target Target,
	opts deployment.AgentEventListOptions,
	onEvent func(ev deployment.AgentEvent, streaming bool) error,
) (int, error) {
	next := opts.MinOffset
	err := api.FollowAgentEvents(
		ctx, target.OrgID, target.ProjectID, target.AgentName, target.SessionID, opts,
		func(ev deployment.AgentEvent) error {
			if !ev.Streaming() {
				next = ev.Offset + 1
				return onEvent(ev, false)
			}
			if err := onEvent(ev, true); err != nil {
				return err
			}
			final, err := api.FollowAgentEvent(
				ctx, target.OrgID, target.ProjectID, target.AgentName, target.SessionID, ev.ID,
				func(update deployment.AgentEvent) error {
					if update.Streaming() {
						return onEvent(update, true)
					}
					return nil
				},
			)
			if err != nil {
				return err
			}
			next = ev.Offset + 1
			return onEvent(*final, false)
		},
	)
	if errors.Is(err, errDone) {
		return next, nil
	}
	return next, err
}

// followError says what a failed follow means; nil when it was simply stopped.
// A stream that ended says how to pick it up, as replay does.
func followError(err error, followAgain string) error {
	switch {
	case errors.Is(err, context.Canceled):
		return nil
	case errors.Is(err, deployment.ErrEventStreamEnded):
		return fmt.Errorf("%w; follow it again with: %s", err, followAgain)
	}
	return err
}

// followEventsCommand is how to keep following a session's events with the same filters and output.
func followEventsCommand(
	target Target,
	opts deployment.AgentEventListOptions,
	outputFlag string,
) string {
	parts := []string{
		"iai agents events list", target.AgentName, target.SessionID,
		"--min-offset", strconv.Itoa(opts.MinOffset),
	}
	for _, filter := range []struct{ flag, value string }{
		{"--kinds", opts.Kinds},
		{"--source", opts.Source},
		{"--trace-id", opts.TraceID},
	} {
		if filter.value != "" {
			parts = append(parts, filter.flag, filter.value)
		}
	}
	parts = append(parts, "--follow")
	if outputFlag != "" {
		parts = append(parts, outputFlag)
	}
	return strings.Join(parts, " ")
}

// followEventCommand is how to keep following one event's updates with the same output.
func followEventCommand(target Target, eventID, outputFlag string) string {
	command := "iai agents events get " + target.AgentName + " " + target.SessionID + " " +
		eventID + " --follow"
	if outputFlag != "" {
		command += " " + outputFlag
	}
	return command
}
