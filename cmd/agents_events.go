package cmd

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/conversation"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	agentEventMessage    string
	agentEventFile       string
	agentEventCreateJSON bool
	agentEventCreateYAML bool

	agentEventMinOffset  int
	agentEventMaxOffset  int
	agentEventLimit      int
	agentEventSource     string
	agentEventKinds      string
	agentEventTraceID    string
	agentEventListFollow bool
	agentEventListJSON   bool
	agentEventListYAML   bool

	agentEventGetFollow bool
	agentEventGetJSON   bool
	agentEventGetYAML   bool
)

var agentEventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event"},
	Short:   "Events in a conversation session on an agent",
	Long: `Read and create the events of a conversation session on a deployed agent.

A conversation is an ordered list of events: the messages each side sends,
status events as the agent works on a reply, tool events and custom events.
Each event has an offset that orders it in the conversation.`,
}

var agentEventCreateCmd = &cobra.Command{
	Use:   "create <agent_name> <session_id>",
	Short: "Send an event to a conversation session",
	Long:  `Send an event to a conversation session on a deployed agent.`,
	Example: `  iai agents events create my-agent 9mtKfiDIaZ --message "Where is my order?"
  iai agents events create my-agent 9mtKfiDIaZ --file event.json
  cat event.json | iai agents events create my-agent 9mtKfiDIaZ --file -`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName, sessionID := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])

		body, err := inputs.BuildEventBody(agentEventMessage, agentEventFile, cmd.InOrStdin())
		if err != nil {
			return err
		}
		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		event, err := deployClient.CreateAgentEvent(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agentName, sessionID, body,
		)
		if err != nil {
			return err
		}
		if agentEventCreateJSON {
			return output.PrintRawJSON(out, event.Raw)
		}
		if agentEventCreateYAML {
			return output.PrintRawYAML(out, event.Raw)
		}
		return output.PrintAgentEvent(out, event)
	},
}

var agentEventListCmd = &cobra.Command{
	Use:     "list <agent_name> <session_id>",
	Aliases: []string{"ls"},
	Short:   "List the events of a conversation session",
	Long:    `List the events of a conversation session on a deployed agent, in offset order.`,
	Example: `  iai agents events list my-agent 9mtKfiDIaZ
  iai agents events list my-agent 9mtKfiDIaZ --kinds message
  iai agents events list my-agent 9mtKfiDIaZ --min-offset 12 --follow
  iai agents events list my-agent 9mtKfiDIaZ --follow --json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName, sessionID := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
		opts := deployment.AgentEventListOptions{
			MinOffset: agentEventMinOffset,
			Limit:     agentEventLimit,
			Source:    agentEventSource,
			Kinds:     agentEventKinds,
			TraceID:   agentEventTraceID,
		}
		if cmd.Flags().Changed("max-offset") {
			opts.MaxOffset = &agentEventMaxOffset
		}

		if agentEventListFollow {
			print := output.EventFollower(out, agentEventListJSON, agentEventListYAML)
			return runFollow(
				cmd,
				agentName,
				sessionID,
				func(ctx context.Context, client *deployment.DeploymentClient, target conversation.Target) error {
					return conversation.FollowEvents(ctx, client, target, opts,
						outputFlag(agentEventListJSON, agentEventListYAML), print)
				},
			)
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		list, err := deployClient.ListAgentEvents(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agentName, sessionID, opts,
		)
		if err != nil {
			return err
		}
		if agentEventListJSON {
			return output.PrintRawJSON(out, list.Raw)
		}
		if agentEventListYAML {
			return output.PrintRawYAML(out, list.Raw)
		}
		return output.PrintAgentEventList(out, list.Items)
	},
}

var agentEventGetCmd = &cobra.Command{
	Use:     "get <agent_name> <session_id> <event_id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get one event of a conversation session",
	Long:    `Get one event of a conversation session on a deployed agent.`,
	Example: `  iai agents events get my-agent 9mtKfiDIaZ 3f9c2a
  iai agents events get my-agent 9mtKfiDIaZ 3f9c2a --follow
  iai agents events get my-agent 9mtKfiDIaZ 3f9c2a --json`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])
		sessionID, eventID := strings.TrimSpace(args[1]), strings.TrimSpace(args[2])

		if agentEventGetFollow {
			print := output.EventFollower(out, agentEventGetJSON, agentEventGetYAML)
			return runFollow(
				cmd,
				agentName,
				sessionID,
				func(ctx context.Context, client *deployment.DeploymentClient, target conversation.Target) error {
					return conversation.FollowEvent(ctx, client, target, eventID,
						outputFlag(agentEventGetJSON, agentEventGetYAML), print)
				},
			)
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		event, err := deployClient.GetAgentEvent(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agentName, sessionID, eventID,
		)
		if err != nil {
			return err
		}
		if agentEventGetJSON {
			return output.PrintRawJSON(out, event.Raw)
		}
		if agentEventGetYAML {
			return output.PrintRawYAML(out, event.Raw)
		}
		return output.PrintAgentEvent(out, event)
	},
}

// outputFlag is the structured output flag in use, repeated in a hint to follow again.
func outputFlag(asJSON, asYAML bool) string {
	switch {
	case asJSON:
		return "--json"
	case asYAML:
		return "--yaml"
	}
	return ""
}

// runFollow resolves the session to follow and runs follow until Ctrl+C; a followed
// stream is one long response, so there is no client timeout, as with logs --follow.
func runFollow(
	cmd *cobra.Command,
	agentName, sessionID string,
	follow func(context.Context, *deployment.DeploymentClient, conversation.Target) error,
) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pCtx, _, deployClient, err := resolveProject(
		ctx, agentOrganization, agentProject, resolveOpts{deployTimeout: 0},
	)
	if err != nil {
		return err
	}
	return follow(ctx, deployClient, conversation.Target{
		OrgID:     pCtx.orgId,
		ProjectID: pCtx.projectId,
		AgentName: agentName,
		SessionID: sessionID,
	})
}

func init() {
	f := agentEventCreateCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.StringVar(&agentEventMessage, "message", "", "Customer message to send; the agent answers it")
	f.StringVar(
		&agentEventFile,
		"file",
		"",
		"Event as JSON, from a file or stdin (-): a message, custom or status event using the platform API fields, including participant.displayName",
	)
	f.BoolVar(&agentEventCreateJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&agentEventCreateYAML, "yaml", false, "Output raw API response as YAML")
	agentEventCreateCmd.MarkFlagsOneRequired("message", "file")
	agentEventCreateCmd.MarkFlagsMutuallyExclusive("message", "file")
	agentEventCreateCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	f = agentEventListCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.IntVar(&agentEventMinOffset, "min-offset", 0, "First offset to return")
	f.IntVar(&agentEventMaxOffset, "max-offset", 0, "Last offset to return; not with --follow")
	f.IntVar(&agentEventLimit, "limit", 0,
		"Return only this many of the newest matching events (1-500); not with --follow")
	indexFlag(agentEventListCmd, "limit", "1-500")
	f.StringVar(&agentEventSource, "source", "",
		"Only events from this source: customer, customerUi, human, "+
			"humanOnBehalfOfAgent, agent or system")
	f.StringVar(&agentEventKinds, "kinds", "",
		"Only these kinds, comma-separated: message, tool, status, custom")
	f.StringVar(&agentEventTraceID, "trace-id", "", "Only events of this trace")
	f.BoolVarP(
		&agentEventListFollow,
		"follow",
		"f",
		false,
		"Stream new events, updating a reply in place while the agent writes it, for up to 10 minutes; "+
			"reconnect to continue; cannot combine with --max-offset or --limit",
	)
	f.BoolVar(&agentEventListJSON, "json", false,
		"Output raw API response as JSON; with --follow, one record per complete event")
	f.BoolVar(&agentEventListYAML, "yaml", false,
		"Output raw API response as YAML; with --follow, one document per complete event")
	agentEventListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	agentEventListCmd.MarkFlagsMutuallyExclusive("follow", "max-offset")
	agentEventListCmd.MarkFlagsMutuallyExclusive("follow", "limit")

	f = agentEventGetCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.BoolVarP(
		&agentEventGetFollow,
		"follow",
		"f",
		false,
		"Stream a message the agent is still writing, updated in place until it is complete, "+
			"for up to 10 minutes; reconnect to continue",
	)
	f.BoolVar(&agentEventGetJSON, "json", false,
		"Output raw API response as JSON; with --follow, the event once it is complete")
	f.BoolVar(&agentEventGetYAML, "yaml", false,
		"Output raw API response as YAML; with --follow, the event once it is complete")
	agentEventGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentEventsCmd.AddCommand(agentEventCreateCmd, agentEventListCmd, agentEventGetCmd)
	agentsCmd.AddCommand(agentEventsCmd)
}
