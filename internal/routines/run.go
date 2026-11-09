// Package routines launches autonomous routines in retained conversation sessions.
package routines

import (
	"context"
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

// Start creates a session when needed and triggers once, retaining the session even on failure.
func Start(
	ctx context.Context,
	client *deployment.DeploymentClient,
	orgID, projectID, agentName, routineID string,
	request deployment.AgentRoutineTriggerRequest,
	stderr io.Writer,
) (*deployment.AgentRoutineRun, error) {
	createdSessionID := ""
	if request.SessionID == "" {
		session, err := client.CreateAgentSession(ctx, orgID, projectID, agentName,
			deployment.AgentSessionCreateRequest{}, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create routine session: %w", err)
		}
		request.SessionID = session.ID
		createdSessionID = session.ID
		fmt.Fprintf(
			stderr,
			"Created session %s on %s; it will be retained.\n",
			session.ID,
			agentName,
		)
	}

	run, err := client.TriggerAgentRoutine(ctx, orgID, projectID, agentName, routineID, request)
	if run != nil {
		if createdSessionID != "" && run.SessionID != createdSessionID {
			fmt.Fprintf(
				stderr,
				"Existing run uses session %s; newly created session %s remains unused.\n",
				run.SessionID,
				createdSessionID,
			)
		}
		return run, err
	}
	return nil, fmt.Errorf(
		"failed to trigger routine in session %s (retained); execution may already have started, do not automatically retry: %w",
		request.SessionID,
		err,
	)
}
