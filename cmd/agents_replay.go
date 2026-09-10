package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/agent"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/replay"
	"github.com/spf13/cobra"
)

var (
	replayDataset     string
	replayScenarios   []string
	replayFile        string
	replayRunID       string
	replayRepeat      int
	replayConcurrency int
	replayTimeout     time.Duration
	replayAgentURL    string
	replayAPIKey      string
	replayJSON        bool
)

var agentReplayCmd = &cobra.Command{
	Use:   "replay <agent_name>",
	Short: "Replay recorded scenarios against an agent and report the verdict",
	Long: `Replay recorded scenarios against a deployed agent and report the verdict.

A scenario is a recorded conversation stored as a dataset item: customer
messages, context variables, the tool results the agent saw, and optionally
what the run must satisfy. Replaying re-runs those inputs with live reasoning,
answers tool calls from the recorded fixtures, then checks the expectations
and writes PASS/FAIL scores to the platform.

The platform runs the replay against the agent and streams its progress back
over one connection, so an agent without an endpoint replays like any other
and the agent's own key is never needed here. Any agent on version 0.15.0 or
later can be replayed; there is nothing to enable. Three modes:

  --dataset D            replay every scenario in the dataset, or only the
                         names given with --scenarios
  --file PATH            replay a local scenario file (YAML or JSON) inline
  --run-id ID            re-attach to a run already started on the agent

Iterations run in parallel on the agent, up to --concurrency across the whole
run. A replay writes a synthetic customer, a session, and variable values to
the target agent and uses your project's quota: prefer a non-production agent.

Output expands every iteration of a scenario when it is the only one in the
run or when it failed or errored; passed scenarios in a multi-scenario run
are one row each. Progress and the pointer to the platform scores go to
stderr, so stdout carries only the verdict (or the --json payload).

Replaying needs permission to run agents in the project; your login is what
authorizes it. The agent's own API key is only needed with --agent-url, which
talks straight to the given address: pass --agent-api-key or set
INTERACTIVE_AGENT_API_KEY for that.

Exit code is 0 when the run finished with a verdict, passed or failed, and 1
when the replay could not run. The verdict itself is in the output; gate on
it in CI with --json and jq -e '.status == "passed"'.`,
	Example: `  iai agents replay agent-chat-dev --dataset replay-chat --scenarios account-lock
  iai agents replay agent-chat-dev --dataset replay-chat --scenarios account-lock --scenarios bonus-misrouted --repeat 3
  iai agents replay agent-chat-dev --dataset replay-chat --repeat 3 --concurrency 16
  iai agents replay agent-chat-dev --file ./account-lock.yaml
  iai agents replay agent-chat-dev --dataset replay-chat --json > run.json
  iai agents replay agent-chat-dev --run-id 9d0c44e1aa52
  iai agents replay my-local-agent --file ./x.yaml --agent-url http://127.0.0.1:8080 --agent-api-key "$AGENT_API_KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := inputs.ReplayInput{
			Dataset:     replayDataset,
			Scenarios:   replayScenarios,
			File:        replayFile,
			RunID:       replayRunID,
			Repeat:      replayRepeat,
			Concurrency: replayConcurrency,
		}
		if err := inputs.ValidateReplayInput(in); err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		pCtx, _, deployClient, err := resolveProject(ctx, agentOrganization, agentProject)
		if err != nil {
			return err
		}

		deps := replay.Deps{
			Deploy: deployClient,
			NewAgent: func(baseURL string, auth agent.Auth, follow bool) replay.AgentAPI {
				return agent.NewClient(baseURL, auth, follow, defaultHTTPTimeout)
			},
			Stdout: cmd.OutOrStdout(),
			Stderr: cmd.ErrOrStderr(),
		}
		return replay.Run(ctx, deps, replay.Options{
			OrgID:     pCtx.orgId,
			ProjectID: pCtx.projectId,
			AgentName: strings.TrimSpace(args[0]),
			Input:     in,
			AgentURL:  replayAgentURL,
			APIKey:    replayAPIKey,
			APIKeyEnv: os.Getenv(replay.APIKeyEnv),
			Timeout:   replayTimeout,
			JSON:      replayJSON,
		})
	},
}

func init() {
	f := agentReplayCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.StringVar(
		&replayDataset,
		"dataset",
		"",
		"Dataset holding the scenarios (required unless --file or --run-id)",
	)
	f.StringArrayVar(
		&replayScenarios,
		"scenarios",
		nil,
		"Only this scenario name from --dataset; repeat the flag once per scenario, run in this order",
	)
	f.StringVar(&replayFile, "file", "", "Local scenario file (YAML or JSON), posted inline")
	f.StringVar(&replayRunID, "run-id", "",
		"Re-attach to a run already started on the agent; no new run is started")
	f.IntVar(&replayRepeat, "repeat", 1, "Iterations per scenario (1-20)")
	f.IntVar(
		&replayConcurrency,
		"concurrency",
		8,
		"In-flight iterations across the whole run (1-32)",
	)
	f.DurationVar(
		&replayTimeout,
		"timeout",
		30*time.Minute,
		"Give up waiting for the verdict after this long",
	)
	f.StringVar(
		&replayAgentURL,
		"agent-url",
		"",
		"Talk straight to an agent at this base URL instead of through the platform (e.g. http://127.0.0.1:8080)",
	)
	f.StringVar(
		&replayAPIKey,
		"agent-api-key",
		"",
		"Bearer for the agent; only used with --agent-url (else INTERACTIVE_AGENT_API_KEY)",
	)
	f.BoolVar(
		&replayJSON,
		"json",
		false,
		"Print the final run payload exactly as the agent returned it; progress still goes to stderr",
	)

	agentReplayCmd.MarkFlagsMutuallyExclusive("dataset", "file", "run-id")
	agentReplayCmd.MarkFlagsMutuallyExclusive("scenarios", "file")
	agentReplayCmd.MarkFlagsMutuallyExclusive("scenarios", "run-id")
	agentReplayCmd.MarkFlagsMutuallyExclusive("repeat", "run-id")
	agentReplayCmd.MarkFlagsMutuallyExclusive("concurrency", "run-id")

	agentsCmd.AddCommand(agentReplayCmd)
}
