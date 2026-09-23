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

// ReplayAPI is what both clients do: start a run, then wait for its verdict.
type ReplayAPI interface {
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
	Deploy ReplayAPI
	// Describe is nil for a local run: there is no release to describe.
	Describe func(context.Context) (*deployment.DescribeAgentResponse, error)
	Stdout   io.Writer
	Stderr   io.Writer
}

type Options struct {
	OrgID, ProjectID, AgentName string
	Input                       inputs.ReplayInput
	// AgentURL is set for a local run, which changes what a refusal means.
	AgentURL string
	Timeout  time.Duration
	JSON     bool
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

	var version string
	if deps.Describe != nil {
		described, err := deps.Describe(ctx)
		if err != nil {
			return fmt.Errorf("failed to describe agent %q: %w", opts.AgentName, err)
		}
		version = described.Version
		fmt.Fprintf(
			deps.Stderr, "%s  %s  rev %d\n",
			opts.AgentName, described.Version, described.Revision,
		)
	} else {
		fmt.Fprintf(deps.Stderr, "%s  local agent at %s\n", opts.AgentName, opts.AgentURL)
	}

	runID := opts.Input.RunID
	if runID == "" {
		resp, err := deps.Deploy.StartReplay(
			ctx, opts.OrgID, opts.ProjectID, opts.AgentName,
			inputs.BuildReplayStartRequest(opts.Input, scenarioBody),
		)
		if err != nil {
			return refusalError(err, opts.AgentName, version, opts.AgentURL)
		}
		output.PrintReplaySkipped(deps.Stderr, resp.Skipped, opts.Input.Scenarios)
		runID = resp.RunID
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	progress := output.NewReplayProgress(deps.Stderr)
	run, err := deps.Deploy.FollowReplay(
		waitCtx, opts.OrgID, opts.ProjectID, opts.AgentName, runID,
		progress.Update,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			// Stop watching quietly, as logs --follow does; the run keeps going.
			fmt.Fprintf(
				deps.Stderr,
				"stopped watching run %s; follow it again with: %s\n",
				runID,
				reattachCommand(opts.AgentName, runID, opts.AgentURL),
			)
		}
		return waitError(err, opts, version, runID)
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

// reattachCommand is how to pick a run back up, reaching the same agent it is on.
func reattachCommand(agentName, runID, agentURL string) string {
	cmd := fmt.Sprintf("iai agents replay %s --run-id %s", agentName, runID)
	if agentURL != "" {
		cmd += " --agent-url " + agentURL
	}
	return cmd
}

// waitError says what a failed wait means; nil when the watch was simply stopped.
func waitError(err error, opts Options, version, runID string) error {
	reattach := reattachCommand(opts.AgentName, runID, opts.AgentURL)
	switch {
	case errors.Is(err, context.Canceled):
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf(
			"gave up waiting after %s; the run continues on the agent, follow it again with: %s",
			opts.Timeout,
			reattach,
		)
	}
	// Only a refusal under 500 is the agent saying no; a 5xx is a hiccup in front of it.
	// Just the 401 arm: a 404 here is an unknown run, not a missing route.
	var refused *deployment.ReplayError
	if errors.As(err, &refused) && refused.Status < http.StatusInternalServerError {
		if refused.Status == http.StatusUnauthorized {
			return refusalError(err, opts.AgentName, version, opts.AgentURL)
		}
		return err
	}
	// Anything else leaves the run going on the agent, so say how to pick it up.
	return fmt.Errorf("%w; the run may still be going, follow it again with: %s", err, reattach)
}

// refusalError rewords the refusals whose fix is on the caller's side and appends
// skipped items to the rest so the agent's own explanation is never lost.
func refusalError(err error, agentName, version, agentURL string) error {
	var re *deployment.ReplayError
	if !errors.As(err, &re) {
		return err
	}
	switch re.Status {
	case http.StatusUnauthorized:
		if agentURL != "" {
			return fmt.Errorf("the agent at %s rejected --agent-api-key", agentURL)
		}
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
