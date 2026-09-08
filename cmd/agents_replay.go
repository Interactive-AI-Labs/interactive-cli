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

The command talks to the agent itself over its public hostname, not to the
platform. Any agent running agent-server 0.15.0 or later serves the replay
routes; there is nothing to enable. Three modes:

  --dataset D            replay every scenario in the dataset, or only the
                         names given with --scenarios
  --file PATH            replay a local scenario file (YAML or JSON) inline
  --run-id ID            re-attach to a run already started on the agent

Iterations run in parallel on the agent, up to --concurrency across the whole
run. A replay writes a synthetic customer, a session, and variable values to
the target agent and spends router quota: prefer a non-production agent.

The agent's API key is taken from --agent-api-key, then INTERACTIVE_AGENT_API_KEY,
then resolved from the agent's own configuration by reading the project
secret it mounts. That last step needs secret-read permission, so CI should
set INTERACTIVE_AGENT_API_KEY instead.

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
			NewAgent: func(baseURL, bearer string) replay.AgentAPI {
				return agent.NewClient(baseURL, bearer, defaultHTTPTimeout)
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
	f.DurationVar(&replayTimeout, "timeout", 30*time.Minute, "Give up polling after this long")
	f.StringVar(&replayAgentURL, "agent-url", "",
		"Agent base URL, overriding the public hostname (e.g. http://127.0.0.1:8080)")
	f.StringVar(
		&replayAPIKey,
		"agent-api-key",
		"",
		"Bearer for the agent (else INTERACTIVE_AGENT_API_KEY, else resolved from the agent's secrets)",
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
