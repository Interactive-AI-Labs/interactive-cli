// Package replay drives one `iai agents replay` invocation: describe the agent,
// start (or re-attach to) a run on it, wait for the verdict, and render it.
package replay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
)

type DeploymentAPI interface {
	DescribeAgent(
		ctx context.Context,
		orgID, projectID, name string,
	) (*deployment.DescribeAgentResponse, error)
	StartReplay(
		ctx context.Context,
		orgID, projectID, agentName string,
		req deployment.ReplayStartRequest,
	) (*deployment.ReplayStartResponse, error)
	FollowReplay(
		ctx context.Context,
		orgID, projectID, agentName, runID string,
		onProgress func(*deployment.ReplayRun),
	) (*deployment.ReplayRun, error)
}

type Deps struct {
	Deploy DeploymentAPI
	Stdout io.Writer
	Stderr io.Writer
}

type Options struct {
	OrgID, ProjectID, AgentName string
	Input                       inputs.ReplayInput
	Timeout                     time.Duration
	JSON                        bool
}

// Run returns nil when the run finished with a verdict, passed or failed, and
// an error when it did not: refused, unreachable, timed out, lost, or finished
// with status error.
func Run(ctx context.Context, deps Deps, opts Options) error {
	// Local validation first, so a bad file fails before any network call.
	var scenarioBody map[string]any
	if opts.Input.File != "" {
		body, err := inputs.LoadScenarioFile(opts.Input.File)
		if err != nil {
			return err
		}
		scenarioBody = body
	}

	described, err := deps.Deploy.DescribeAgent(ctx, opts.OrgID, opts.ProjectID, opts.AgentName)
	if err != nil {
		return fmt.Errorf("failed to describe agent %q: %w", opts.AgentName, err)
	}
	output.PrintReplayTarget(deps.Stderr, opts.AgentName, described.Version, described.Revision)

	runID := opts.Input.RunID
	if runID == "" {
		resp, err := deps.Deploy.StartReplay(
			ctx, opts.OrgID, opts.ProjectID, opts.AgentName,
			deployment.ReplayStartRequest{
				Dataset:      opts.Input.Dataset,
				Scenarios:    opts.Input.Scenarios,
				ScenarioBody: scenarioBody,
				Repeat:       opts.Input.Repeat,
				Concurrency:  opts.Input.Concurrency,
			},
		)
		if err != nil {
			return startError(err, opts.AgentName, described.Version)
		}
		output.PrintReplaySkipped(deps.Stderr, resp.Skipped, opts.Input.Scenarios)
		runID = resp.RunID
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	lastFinished := -1
	run, err := deps.Deploy.FollowReplay(
		waitCtx, opts.OrgID, opts.ProjectID, opts.AgentName, runID,
		func(r *deployment.ReplayRun) {
			if finished := countFinished(r); finished != lastFinished && len(r.Batches) > 0 {
				lastFinished = finished
				output.PrintReplayProgress(deps.Stderr, finished, len(r.Batches))
			}
		},
	)
	if err != nil {
		reattach := fmt.Sprintf("iai agents replay %s --run-id %s", opts.AgentName, runID)
		switch {
		case errors.Is(err, context.Canceled):
			// Stop watching quietly, as logs --follow does; the run keeps going.
			fmt.Fprintf(
				deps.Stderr,
				"stopped watching run %s; follow it again with: %s\n",
				runID,
				reattach,
			)
			return nil
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf(
				"gave up waiting after %s; the run continues on the agent, follow it again with: %s",
				opts.Timeout,
				reattach,
			)
		case errors.Is(err, deployment.ErrReplayStreamEnded):
			return fmt.Errorf(
				"run %s may still be running on the agent; follow it again with: %s",
				runID, reattach,
			)
		}
		return err
	}

	if opts.JSON {
		if _, err := deps.Stdout.Write(append(run.Raw, '\n')); err != nil {
			return err
		}
	} else if err := output.PrintReplayRun(deps.Stdout, run, opts.Input.Scenarios); err != nil {
		return err
	}
	output.PrintReplayPointer(deps.Stderr, run)

	if run.Status == deployment.ReplayStatusError {
		return fmt.Errorf("replay %s could not be completed: %s", runID, errorSummary(run))
	}
	return nil
}

func countFinished(r *deployment.ReplayRun) int {
	n := 0
	for _, b := range r.Batches {
		if b.Status != deployment.ReplayStatusRunning {
			n++
		}
	}
	return n
}

// startError rewords the two refusals whose fix is on the caller's side and appends
// skipped items to the rest so the agent's own explanation is never lost.
func startError(err error, agentName, version string) error {
	var re *deployment.ReplayError
	if !errors.As(err, &re) {
		return err
	}
	switch re.Status {
	case http.StatusUnauthorized:
		return fmt.Errorf("not authorized to replay %s; check you are logged in", agentName)
	case http.StatusNotFound:
		return fmt.Errorf(
			"agent %s (%s) does not support replay; it needs version 0.15.0 or later",
			agentName, version,
		)
	}
	if len(re.Skipped) == 0 {
		return err
	}
	var b strings.Builder
	b.WriteString(re.Error())
	for _, s := range re.Skipped {
		fmt.Fprintf(&b, "\n  skipped %s: %s", s.ID, s.Reason)
	}
	return errors.New(b.String())
}

func errorSummary(run *deployment.ReplayRun) string {
	if run.Error != "" {
		return run.Error
	}
	for _, b := range run.Batches {
		if b.Error != "" {
			return b.Error
		}
		for _, it := range b.Iterations {
			if it.Error != "" {
				return it.Error
			}
		}
	}
	return "see the errored iterations above"
}
