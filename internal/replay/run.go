// Package replay drives one `iai agents replay` invocation: describe the agent,
// start (or re-attach to) a run on it, wait for the verdict, and render it.
package replay

import (
	"cmp"
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

// APIKeyEnv supplies the agent's bearer for --agent-url runs; the default path needs no key.
const APIKeyEnv = "INTERACTIVE_AGENT_API_KEY"

type DeploymentAPI interface {
	// ReplayEndpoint is where the platform serves one agent's replay routes.
	ReplayEndpoint(orgID, projectID, agentName string) (string, func(*http.Request) error)
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
	// NewAgent builds the replay client; follow is true when the endpoint streams a run's status.
	NewAgent func(baseURL string, auth agent.Auth, follow bool) AgentAPI
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

	// Only --agent-url needs a key here; through the platform the agent's key never leaves it.
	var baseURL string
	var auth agent.Auth
	target := "via platform"
	if opts.AgentURL != "" {
		bearer := cmp.Or(opts.APIKey, opts.APIKeyEnv)
		if bearer == "" {
			return fmt.Errorf(
				"--agent-url talks straight to the agent, which authenticates the call itself: "+
					"pass --agent-api-key or set %s",
				APIKeyEnv,
			)
		}
		baseURL, target = opts.AgentURL, opts.AgentURL
		auth = func(req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+bearer)
			return nil
		}
	} else {
		baseURL, auth = deps.Deploy.ReplayEndpoint(opts.OrgID, opts.ProjectID, opts.AgentName)
	}
	output.PrintReplayTarget(
		deps.Stderr,
		opts.AgentName,
		described.Version,
		described.Revision,
		target,
	)

	api := deps.NewAgent(baseURL, auth, opts.AgentURL == "")

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
			return startError(err, opts.AgentName, described.Version, opts.AgentURL != "")
		}
		output.PrintReplaySkipped(deps.Stderr, resp.Skipped, opts.Input.Scenarios)
		runID = resp.RunID
	}

	lastFinished := -1
	run, err := api.Wait(ctx, runID, agent.WaitOptions{
		Interval:    pollInterval,
		Timeout:     opts.Timeout,
		NotFoundCap: notFoundCap,
		OnRetry: func(err error) {
			output.PrintReplayRetry(deps.Stderr, err, notFoundCap.String())
		},
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
		case errors.Is(err, agent.ErrStreamEnded):
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

// startError rewords the two refusals whose fix is on the caller's side and appends
// skipped items to the rest so the agent's own explanation is never lost.
func startError(err error, agentName, version string, direct bool) error {
	var ae *agent.Error
	if !errors.As(err, &ae) {
		return err
	}
	switch ae.Status {
	case http.StatusUnauthorized:
		if direct {
			return fmt.Errorf(
				"the agent rejected --agent-api-key; check it matches the key the agent expects",
			)
		}
		return fmt.Errorf("not authorized to replay %s; check you are logged in", agentName)
	case http.StatusNotFound:
		return fmt.Errorf(
			"agent %s (%s) does not support replay; it needs version 0.15.0 or later",
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
