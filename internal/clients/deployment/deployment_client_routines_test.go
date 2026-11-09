package deployment

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAgentRoutineTriggerPath(t *testing.T) {
	tests := []struct {
		name, org, project, agent, routine, want string
	}{
		{
			"simple",
			"o",
			"p",
			"a",
			"refund",
			"/v1/organizations/o/projects/p/agents/a/routines/refund/trigger",
		},
		{
			"folders",
			"o",
			"p",
			"a",
			"billing/orders/refund",
			"/v1/organizations/o/projects/p/agents/a/routines/billing%2Forders%2Frefund/trigger",
		},
		{
			"escaped identifiers",
			"org name",
			"p?",
			"a#",
			"bill%2F/refund?x=1#end",
			"/v1/organizations/org%20name/projects/p%3F/agents/a%23/routines/bill%252F%2Frefund%3Fx=1%23end/trigger",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &DeploymentClient{hostname: "http://localhost:8000"}
			req, err := client.newRequest(context.Background(), http.MethodPost,
				agentRoutineTriggerPath(tt.org, tt.project, tt.agent, tt.routine))
			if err != nil {
				t.Fatal(err)
			}
			if got := req.URL.RequestURI(); got != tt.want {
				t.Errorf("request path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentRoutineTriggerRequest(t *testing.T) {
	tests := []struct {
		name    string
		request AgentRoutineTriggerRequest
		want    string
	}{
		{
			name: "required fields with exact input",
			request: AgentRoutineTriggerRequest{
				Input:       json.RawMessage(`{"n":9007199254740993}`),
				CallbackURL: "https://trusted.example/results",
				SessionID:   "s1",
			},
			want: `{"input":{"n":9007199254740993},"callbackUrl":"https://trusted.example/results","sessionId":"s1"}`,
		},
		{
			name: "idempotency key",
			request: AgentRoutineTriggerRequest{
				Input:          json.RawMessage(`{}`),
				CallbackURL:    "https://trusted.example/results",
				SessionID:      "s1",
				IdempotencyKey: "key-1",
			},
			want: `{"input":{},"callbackUrl":"https://trusted.example/results","sessionId":"s1","idempotencyKey":"key-1"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("body = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDecodeAgentRoutineResponse(t *testing.T) {
	const receipt = `{"runId":"run_1","routineId":"billing/refund","status":"accepted","sessionId":"original","createdAt":"2026-11-09T10:00:00Z","future":true}`
	wantRun := &AgentRoutineRun{
		RunID:     "run_1",
		RoutineID: "billing/refund",
		Status:    "accepted",
		SessionID: "original",
		CreatedAt: "2026-11-09T10:00:00Z",
		Raw:       json.RawMessage(receipt),
	}
	tests := []struct {
		name    string
		status  int
		body    string
		want    *AgentRoutineRun
		wantErr string
	}{
		{"accepted raw receipt retained", 202, receipt, wantRun, ""},
		{
			"conflict retains original receipt",
			409,
			receipt,
			wantRun,
			"agent returned 409: idempotency key already used; no new run started",
		},
		{
			"invalid input",
			400,
			`{"detail":"input validation failed"}`,
			nil,
			"agent returned 400: input validation failed",
		},
		{
			"not found",
			404,
			`{"detail":"routine not found"}`,
			nil,
			"agent returned 404: routine not found",
		},
		{
			"operator failure",
			424,
			`{"message":"Agent request could not be completed"}`,
			nil,
			"agent returned 424: Agent request could not be completed",
		},
		{"null receipt", 202, `null`, nil, "routine receipt is missing required fields"},
		{
			"incomplete receipt",
			202,
			`{"runId":"run_1"}`,
			nil,
			"routine receipt is missing required fields",
		},
		{
			"malformed receipt",
			202,
			`{`,
			nil,
			"failed to decode routine receipt: unexpected end of JSON input",
		},
		{
			"unrelated conflict is not a receipt",
			409,
			`{"message":"conflict"}`,
			nil,
			"routine receipt is missing required fields",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeAgentRoutineResponse(tt.status, []byte(tt.body))
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Fatalf("error = %q, want %q", gotErr, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("receipt (-want +got):\n%s", diff)
			}
		})
	}
}
