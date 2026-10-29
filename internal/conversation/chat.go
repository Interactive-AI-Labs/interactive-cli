package conversation

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
)

type ChatDeps struct {
	API    *deployment.DeploymentClient
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type ChatOptions struct {
	OrgID, ProjectID, AgentName string
	// SessionID resumes a conversation; empty starts one.
	SessionID  string
	CustomerID string
	Title      string
}

// historyMessages is how many of a resumed conversation's latest messages are shown.
const historyMessages = 50

// turnEnded is the status after which the agent has nothing more to add to a reply.
var turnEnded = map[string]bool{"ready": true, "error": true, "cancelled": true}

// workingNotice is what the agent is doing while a status says it is still on the turn.
var workingNotice = map[string]string{
	"acknowledged": "thinking…",
	"processing":   "thinking…",
	"typing":       "typing…",
}

// Chat sends each line read from Stdin as a customer message and shows the agent's
// reply as it is written, until Stdin ends or ctx is cancelled.
func Chat(ctx context.Context, deps ChatDeps, opts ChatOptions) error {
	printer := output.NewConversationPrinter(deps.Stdout, map[string]string{
		deployment.EventSourceCustomer: "you",
		deployment.EventSourceAgent:    opts.AgentName,
	})
	target, err := chatTarget(ctx, deps, opts, printer)
	if err != nil {
		return err
	}

	lines := readLines(deps.Stdin)
	for {
		if printer.TTY() {
			fmt.Fprint(deps.Stdout, output.ChatPrompt())
		}
		var line string
		var ok bool
		select {
		case <-ctx.Done():
			fmt.Fprintln(deps.Stdout)
			return nil
		case line, ok = <-lines:
		}
		if !ok {
			fmt.Fprintln(deps.Stdout)
			return nil
		}
		message := strings.TrimSpace(line)
		if message == "" {
			continue
		}

		sent, err := deps.API.CreateAgentEvent(
			ctx, target.OrgID, target.ProjectID, target.AgentName, target.SessionID,
			inputs.MessageEvent(message),
		)
		if err == nil {
			err = reply(ctx, deps.API, target, sent.Offset+1, printer)
		}
		if ctx.Err() != nil {
			fmt.Fprintln(deps.Stdout)
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// chatTarget creates the session, or shows the latest messages of a resumed one.
func chatTarget(
	ctx context.Context,
	deps ChatDeps,
	opts ChatOptions,
	printer *output.ConversationPrinter,
) (Target, error) {
	target := Target{
		OrgID:     opts.OrgID,
		ProjectID: opts.ProjectID,
		AgentName: opts.AgentName,
		SessionID: opts.SessionID,
	}
	if target.SessionID == "" {
		session, err := deps.API.CreateAgentSession(
			ctx, opts.OrgID, opts.ProjectID, opts.AgentName,
			deployment.AgentSessionCreateRequest{CustomerID: opts.CustomerID, Title: opts.Title},
			false,
		)
		if err != nil {
			return Target{}, err
		}
		target.SessionID = session.ID
	}

	if printer.TTY() {
		fmt.Fprint(deps.Stdout, output.ChatBanner(opts.AgentName, target.SessionID))
	} else {
		fmt.Fprintf(deps.Stderr, "session %s on %s; resume with --session-id %s\n",
			target.SessionID, opts.AgentName, target.SessionID)
	}
	if opts.SessionID == "" {
		return target, nil
	}

	history, err := deps.API.ListAgentEvents(
		ctx, opts.OrgID, opts.ProjectID, opts.AgentName, opts.SessionID,
		deployment.AgentEventListOptions{
			Limit: historyMessages,
			Kinds: deployment.EventKindMessage,
		},
	)
	if err != nil {
		return Target{}, err
	}
	for _, ev := range history.Items {
		printer.Print(ev, false)
	}
	return target, nil
}

// reply shows what the agent does with the message at the offset before from,
// until it reports the turn over.
func reply(
	ctx context.Context,
	api *deployment.DeploymentClient,
	target Target,
	from int,
	printer *output.ConversationPrinter,
) error {
	opts := deployment.AgentEventListOptions{
		MinOffset: from,
		Kinds:     deployment.EventKindMessage + "," + deployment.EventKindStatus,
	}
	printer.Notice(workingNotice["processing"])
	next, err := follow(
		ctx,
		api,
		target,
		opts,
		func(ev deployment.AgentEvent, streaming bool) error {
			if status := ev.Status(); status != "" {
				if turnEnded[status] {
					printer.Clear()
					if status != "ready" {
						printer.Print(ev, false)
					}
					return errDone
				}
				if notice, ok := workingNotice[status]; ok {
					printer.Notice(notice)
				}
				return nil
			}
			printer.Print(ev, streaming)
			return nil
		},
	)
	opts.MinOffset = next
	return followError(err, followEventsCommand(target, opts, ""))
}

func readLines(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()
	return lines
}
