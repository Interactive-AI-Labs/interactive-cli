// Package replay drives one `iai agents replay` invocation: describe the
// agent, resolve its bearer, start (or re-attach to) a run on the agent, poll
// it to completion, and render the verdict.
package replay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/agent"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
)

const (
	pollInterval = 10 * time.Second
	notFoundCap  = 5 * time.Minute
)

type DeploymentAPI interface {
	SecretReader
	DescribeAgent(
		ctx context.Context,
		orgID, projectID, name string,
	) (*deployment.DescribeAgentResponse, error)
}

type AgentAPI interface {
	StartReplay(ctx context.Context, req agent.StartRequest) (*agent.StartResponse, error)
	Wait(ctx context.Context, runID string, opts agent.WaitOptions) (*agent.Run, error)
}

type Deps struct {
	Deploy DeploymentAPI
	// NewAgent builds the agent client once the base URL and bearer are known.
	NewAgent func(baseURL, bearer string) AgentAPI
	Stdout   io.Writer
	Stderr   io.Writer
}

type Options struct {
	OrgID, ProjectID, AgentName string
	Input                       inputs.ReplayInput
	AgentURL                    string
	APIKey, APIKeyEnv           string
	Timeout                     time.Duration
	JSON                        bool
}

// Run returns nil when the run finished with a verdict, passed or failed, and
// an error when it did not: refused, unreachable, timed out, lost, interrupted,
// or finished with status error.
func Run(ctx context.Context, deps Deps, opts Options) error {
	described, err := deps.Deploy.DescribeAgent(ctx, opts.OrgID, opts.ProjectID, opts.AgentName)
	if err != nil {
		return fmt.Errorf("failed to describe agent %q: %w", opts.AgentName, err)
	}

	baseURL := opts.AgentURL
	if baseURL == "" {
		if described.Endpoint == "" {
			return fmt.Errorf(
				"agent %s has no public endpoint; run `iai agents port-forward %s --local-port 8080` "+
					"in another shell and pass --agent-url http://127.0.0.1:8080, "+
					"or expose it with `iai agents update %s --endpoint`",
				opts.AgentName,
				opts.AgentName,
				opts.AgentName,
			)
		}
		baseURL = "https://" + described.Endpoint
	}
	output.PrintReplayTarget(
		deps.Stderr,
		opts.AgentName,
		described.Version,
		described.Revision,
		baseURL,
	)

	var scenarioBody map[string]any
	if opts.Input.File != "" {
		if scenarioBody, err = inputs.LoadScenarioFile(opts.Input.File); err != nil {
			return err
		}
	}

	bearer, err := ResolveBearer(
		ctx,
		deps.Deploy,
		opts.OrgID,
		opts.ProjectID,
		described,
		opts.APIKey,
		opts.APIKeyEnv,
		deps.Stderr,
	)
	if err != nil {
		return err
	}
	api := deps.NewAgent(baseURL, bearer)

	runID := opts.Input.RunID
	if runID == "" {
		resp, err := api.StartReplay(ctx, agent.StartRequest{
			Dataset:      opts.Input.Dataset,
			Scenarios:    opts.Input.Scenarios,
			ScenarioBody: scenarioBody,
			Repeat:       opts.Input.Repeat,
			Concurrency:  opts.Input.Concurrency,
		})
		if err != nil {
			return startError(err, opts.AgentName, described.Version)
		}
		output.PrintReplaySkipped(deps.Stderr, resp.Skipped, opts.Input.Scenarios)
		runID = resp.RunID
	}

	lastFinished := -1
	run, err := api.Wait(ctx, runID, agent.WaitOptions{
		Interval:    pollInterval,
		Timeout:     opts.Timeout,
		NotFoundCap: notFoundCap,
		OnProgress: func(r *agent.Run) {
			if finished := countFinished(r); finished != lastFinished && len(r.Batches) > 0 {
				lastFinished = finished
				output.PrintReplayProgress(deps.Stderr, finished, len(r.Batches))
			}
		},
	})
	if err != nil {
		reattach := fmt.Sprintf("iai agents replay %s --run-id %s", opts.AgentName, runID)
		switch {
		case errors.Is(err, context.Canceled):
			return fmt.Errorf(
				"interrupted; the run continues on the agent, re-attach with: %s",
				reattach,
			)
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf(
				"gave up polling after %s; the run continues on the agent, re-attach with: %s",
				opts.Timeout, reattach,
			)
		}
		return err
	}

	if opts.JSON {
		if _, err := deps.Stdout.Write(append(run.Raw, '\n')); err != nil {
			return err
		}
	} else if err := output.PrintReplayRun(deps.Stdout, deps.Stderr, run); err != nil {
		return err
	}

	if run.Status == agent.StatusError {
		return fmt.Errorf("replay %s could not be completed: %s", runID, errorSummary(run))
	}
	return nil
}

func countFinished(r *agent.Run) int {
	n := 0
	for _, b := range r.Batches {
		if b.Status != agent.StatusRunning {
			n++
		}
	}
	return n
}

// startError rewords the two refusals whose fix is on the CLI side and appends
// skipped items to the rest so the agent's own explanation is never lost.
func startError(err error, agentName, version string) error {
	var ae *agent.Error
	if !errors.As(err, &ae) {
		return err
	}
	switch ae.Status {
	case http.StatusUnauthorized:
		return fmt.Errorf(
			"the agent rejected the api key; check AGENT_API_KEY in the agent's secret, or pass --agent-api-key",
		)
	case http.StatusNotFound:
		return fmt.Errorf(
			"agent %s (%s) has no /replays route; replay needs agent-server 0.15.0 or later",
			agentName, version,
		)
	}
	if len(ae.Skipped) == 0 {
		return err
	}
	var b strings.Builder
	b.WriteString(ae.Error())
	for _, s := range ae.Skipped {
		fmt.Fprintf(&b, "\n  skipped %s: %s", s.ID, s.Reason)
	}
	return errors.New(b.String())
}

func errorSummary(run *agent.Run) string {
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
