package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

type AgentRoutineTriggerRequest struct {
	Input          json.RawMessage `json:"input"`
	CallbackURL    string          `json:"callbackUrl"`
	SessionID      string          `json:"sessionId"`
	IdempotencyKey string          `json:"idempotencyKey,omitempty"`
}

type AgentRoutineRun struct {
	RunID     string          `json:"runId"`
	RoutineID string          `json:"routineId"`
	Status    string          `json:"status"`
	SessionID string          `json:"sessionId"`
	CreatedAt string          `json:"createdAt"`
	Raw       json.RawMessage `json:"-"`
}

func agentRoutineTriggerPath(orgID, projectID, agentName, routineID string) string {
	return fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/agents/%s/routines/%s/trigger",
		url.PathEscape(orgID), url.PathEscape(projectID),
		url.PathEscape(agentName), url.PathEscape(routineID),
	)
}

// TriggerAgentRoutine returns the existing receipt alongside a conflict error for duplicate keys.
func (c *DeploymentClient) TriggerAgentRoutine(
	ctx context.Context,
	orgID, projectID, agentName, routineID string,
	request AgentRoutineTriggerRequest,
) (*AgentRoutineRun, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode routine trigger: %w", err)
	}
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		agentRoutineTriggerPath(orgID, projectID, agentName, routineID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create routine trigger request: %w", err)
	}
	clients.SetJSONBody(req, body)
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("routine trigger request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read routine trigger response: %w", err)
	}
	return decodeAgentRoutineResponse(resp.StatusCode, raw)
}

func decodeAgentRoutineResponse(status int, raw []byte) (*AgentRoutineRun, error) {
	if status != http.StatusAccepted && status != http.StatusConflict {
		return nil, DecodeAgentError(status, raw)
	}
	var run AgentRoutineRun
	if err := json.Unmarshal(raw, &run); err != nil {
		return nil, fmt.Errorf("failed to decode routine receipt: %w", err)
	}
	if run.RunID == "" || run.RoutineID == "" || run.Status == "" || run.SessionID == "" ||
		run.CreatedAt == "" {
		return nil, fmt.Errorf("routine receipt is missing required fields")
	}
	run.Raw = raw
	if status == http.StatusConflict {
		return &run, &AgentError{
			Status: status,
			Detail: "idempotency key already used; no new run started",
		}
	}
	return &run, nil
}
