package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/routines"
	"github.com/spf13/cobra"
)

var (
	agentRoutineInputFile      string
	agentRoutineCallbackURL    string
	agentRoutineSessionID      string
	agentRoutineIdempotencyKey string
	agentRoutineJSON           bool
	agentRoutineYAML           bool
)

var agentRoutinesCmd = &cobra.Command{
	Use:     "routines",
	Aliases: []string{"routine"},
	Short:   "Execute autonomous routines on a deployed agent",
	Long: `Execute autonomous routines already loaded by a deployed agent.

Use 'iai routines' to manage routine definitions and 'iai agents events' to
inspect execution events.`,
}

var agentRoutineRunCmd = &cobra.Command{
	Use:   "run <agent_name> <routine_id>",
	Short: "Launch an existing autonomous routine",
	Long: `Launch an existing autonomous routine and return its receipt without waiting
for completion. The input is data for the routine, not a routine definition.

The session is retained so you can inspect tool activity and output with
'iai agents events list'. Session events do not report every final failure;
the completion callback remains the authoritative outcome.

A duplicate idempotency key prints the original receipt and exits with an error.
No requests are automatically retried. Any session created before a failed or
duplicate trigger remains available; its ID is printed to stderr.`,
	Example: `  iai agents routines run my-agent billing/refund --input-file input.json --callback-url https://trusted.example/results
  iai agents routines run my-agent billing/refund --input-file input.json --callback-url https://trusted.example/results --session-id sess_123
  cat input.json | iai agents routines run my-agent billing/refund --input-file - --callback-url https://trusted.example/results --idempotency-key refund-order-123 --json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName, routineID := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
		if err := inputs.ValidateRoutineRun(
			agentName,
			routineID,
			agentRoutineCallbackURL,
		); err != nil {
			return err
		}
		if cmd.Flags().Changed("session-id") && strings.TrimSpace(agentRoutineSessionID) == "" {
			return fmt.Errorf("--session-id must not be empty")
		}
		input, err := inputs.ReadRoutineInput(agentRoutineInputFile, cmd.InOrStdin())
		if err != nil {
			return err
		}

		pCtx, _, client, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		run, triggerErr := routines.Start(cmd.Context(), client, pCtx.orgId, pCtx.projectId,
			agentName, routineID, deployment.AgentRoutineTriggerRequest{
				Input: input, CallbackURL: agentRoutineCallbackURL,
				SessionID: agentRoutineSessionID, IdempotencyKey: agentRoutineIdempotencyKey,
			}, cmd.ErrOrStderr())
		if run == nil {
			return triggerErr
		}

		out := cmd.OutOrStdout()
		switch {
		case agentRoutineJSON:
			err = output.PrintRawJSON(out, run.Raw)
		case agentRoutineYAML:
			err = output.PrintRawYAML(out, run.Raw)
		default:
			err = output.PrintAgentRoutineRun(out, run)
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Follow execution:\n  %s\n",
			output.RoutineFollowCommand(agentName, run.SessionID, pCtx.orgName, pCtx.projectName,
				hostnameOverride(hostname, cmd.Flag("hostname").DefValue),
				hostnameOverride(deploymentHostname, cmd.Flag("deployment-hostname").DefValue)))
		return triggerErr
	},
}

// hostnameOverride accounts for the root command adding HTTPS to schemeless defaults.
func hostnameOverride(value, defaultValue string) string {
	if value == defaultValue || value == "https://"+defaultValue {
		return ""
	}
	return value
}

func init() {
	f := agentRoutineRunCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.StringVar(&agentRoutineInputFile, "input-file", "",
		"Routine input as a JSON object, from a file or stdin (-); not a routine definition")
	f.StringVar(
		&agentRoutineCallbackURL,
		"callback-url",
		"",
		"Trusted HTTP(S) URL for the completion result; required by the agent, allowed by the routine, and receives the agent's bearer credential",
	)
	f.StringVar(&agentRoutineSessionID, "session-id", "",
		"Use this existing session; when omitted, create a new retained session without a greeting")
	f.StringVar(
		&agentRoutineIdempotencyKey,
		"idempotency-key",
		"",
		"Key to prevent duplicate execution; duplicates print the existing run and exit with an error; deduplication does not survive agent restarts",
	)
	f.BoolVar(
		&agentRoutineJSON,
		"json",
		false,
		"Output raw receipt as JSON; session notices and follow instructions go to stderr",
	)
	f.BoolVar(
		&agentRoutineYAML,
		"yaml",
		false,
		"Output raw receipt as YAML; session notices and follow instructions go to stderr",
	)
	_ = agentRoutineRunCmd.MarkFlagRequired("input-file")
	_ = agentRoutineRunCmd.MarkFlagRequired("callback-url")
	agentRoutineRunCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentRoutinesCmd.AddCommand(agentRoutineRunCmd)
	agentsCmd.AddCommand(agentRoutinesCmd)
}
