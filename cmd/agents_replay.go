package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	replayJSON        bool
)

var agentReplayCmd = &cobra.Command{
	Use:   "replay <agent_name>",
	Short: "Test an agent by replaying recorded scenarios and report the verdict",
	Long: `Test a deployed agent by replaying recorded scenarios and report the verdict.

A scenario is a dataset item: a recorded conversation's customer messages,
context and tool results, plus what the run must satisfy. The platform re-runs
it against the agent with live reasoning, answers tool calls from the recorded
fixtures, checks the expectations and writes PASS/FAIL scores. Replay a whole
dataset (--dataset, optionally narrowed with --scenarios), a local scenario
file (--file), or re-attach to a run in flight (--run-id).

A replay writes a synthetic customer and session to the agent and uses your
project's quota: prefer a non-production agent. Progress goes to stderr and
only the verdict to stdout. Exit code is 0 for any verdict, 1 when the replay
could not run; gate in CI with --json and jq -e '.status == "passed"'.`,
	Example: `  iai agents replay agent-chat-dev --dataset replay-chat --scenarios account-lock
  iai agents replay agent-chat-dev --dataset replay-chat --scenarios account-lock --scenarios bonus-misrouted --repeat 3
  iai agents replay agent-chat-dev --dataset replay-chat --repeat 3 --concurrency 16
  iai agents replay agent-chat-dev --file ./account-lock.yaml
  iai agents replay agent-chat-dev --dataset replay-chat --json > run.json
  iai agents replay agent-chat-dev --run-id 9d0c44e1aa52`,
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
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// A followed run is one long response, so no client-wide timeout; as logs --follow.
		pCtx, _, deployClient, err := resolveProject(
			ctx, agentOrganization, agentProject, resolveOpts{deployTimeout: 0},
		)
		if err != nil {
			return err
		}

		deps := replay.Deps{
			Deploy: deployClient,
			Stdout: cmd.OutOrStdout(),
			Stderr: cmd.ErrOrStderr(),
		}
		return replay.Run(ctx, deps, replay.Options{
			OrgID:     pCtx.orgId,
			ProjectID: pCtx.projectId,
			AgentName: strings.TrimSpace(args[0]),
			Input:     in,
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
