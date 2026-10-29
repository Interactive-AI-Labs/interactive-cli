package deployment

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAgentSessionListQuery(t *testing.T) {
	tests := []struct {
		name string
		opts AgentSessionListOptions
		want url.Values
	}{
		{name: "defaults send nothing", want: url.Values{}},
		{
			name: "every filter",
			opts: AgentSessionListOptions{
				CustomerID: "c1",
				Labels:     []string{"vip", "test"},
				Limit:      10,
				Cursor:     "a+/=",
				Sort:       "asc",
			},
			want: url.Values{
				"customerId": {"c1"}, "labels": {"vip", "test"}, "limit": {"10"},
				"cursor": {"a+/="}, "sort": {"asc"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, AgentSessionListQuery(tt.opts)); diff != "" {
				t.Errorf("query (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAgentEventListQuery(t *testing.T) {
	zero, twenty := 0, 20
	tests := []struct {
		name string
		opts AgentEventListOptions
		want url.Values
	}{
		{
			name: "unset bounds",
			opts: AgentEventListOptions{},
			want: url.Values{"minOffset": {"0"}},
		},
		{
			name: "zero max offset is a bound",
			opts: AgentEventListOptions{MaxOffset: &zero},
			want: url.Values{"minOffset": {"0"}, "maxOffset": {"0"}},
		},
		{
			name: "every filter",
			opts: AgentEventListOptions{
				MinOffset: 4,
				MaxOffset: &twenty,
				Limit:     5,
				Source:    "agent",
				Kinds:     "message,status",
				TraceID:   "t1",
			},
			want: url.Values{
				"minOffset": {"4"}, "maxOffset": {"20"}, "limit": {"5"},
				"source": {"agent"}, "kinds": {"message,status"}, "traceId": {"t1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, AgentEventListQuery(tt.opts)); diff != "" {
				t.Errorf("query (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecodeAgentEventLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantEvent *AgentEvent
		wantDone  bool
		wantErr   string
	}{
		{name: "heartbeat", line: `{"heartbeat":true}`},
		{name: "timeout", line: `{"timeout":true}`, wantDone: true},
		{
			name:    "error",
			line:    `{"error":"Agent request could not be completed"}`,
			wantErr: "Agent request could not be completed",
		},
		{
			name: "event",
			line: `{"event":{"id":"e1","kind":"status","source":"agent","offset":3,"data":{"status":"ready"}}}`,
			wantEvent: &AgentEvent{
				ID:     "e1",
				Kind:   "status",
				Source: "agent",
				Offset: 3,
				Data:   json.RawMessage(`{"status":"ready"}`),
				Raw: json.RawMessage(
					`{"id":"e1","kind":"status","source":"agent","offset":3,"data":{"status":"ready"}}`,
				),
			},
		},
		{
			name:    "not JSON",
			line:    `nope`,
			wantErr: "failed to decode event stream: invalid character 'o' in literal null (expecting 'u')",
		},
		{
			name:    "invalid event",
			line:    `{"event":[]}`,
			wantErr: "failed to decode event: json: cannot unmarshal array into Go value of type deployment.AgentEvent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, done, err := DecodeAgentEventLine([]byte(tt.line))
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Fatalf("error = %q, want %q", gotErr, tt.wantErr)
			}
			if done != tt.wantDone {
				t.Errorf("done = %v, want %v", done, tt.wantDone)
			}
			if diff := cmp.Diff(tt.wantEvent, event); diff != "" {
				t.Errorf("event (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAgentEventStreaming(t *testing.T) {
	tests := []struct {
		name  string
		event AgentEvent
		want  bool
	}{
		{
			name: "block message",
			event: AgentEvent{
				Kind: "message",
				Data: json.RawMessage(`{"message":"hi","status":"completed"}`),
			},
		},
		{
			name: "streaming started",
			event: AgentEvent{
				Kind: "message",
				Data: json.RawMessage(`{"message":"","status":"streaming"}`),
			},
			want: true,
		},
		{
			name: "streaming",
			event: AgentEvent{
				Kind: "message",
				Data: json.RawMessage(`{"message":"hi","status":"streaming"}`),
			},
			want: true,
		},
		{
			name: "complete",
			event: AgentEvent{
				Kind: "message",
				Data: json.RawMessage(`{"message":"hi","status":"completed"}`),
			},
		},
		{
			name: "status",
			event: AgentEvent{
				Kind: "status",
				Data: json.RawMessage(`{"status":"typing"}`),
			},
		},
		{name: "invalid data", event: AgentEvent{Kind: "message", Data: json.RawMessage(`[]`)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.Streaming(); got != tt.want {
				t.Errorf("streaming = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentEventStatus(t *testing.T) {
	tests := []struct {
		name  string
		event AgentEvent
		want  string
	}{
		{
			name: "status",
			event: AgentEvent{
				Kind: "status",
				Data: json.RawMessage(`{"status":"processing","data":{}}`),
			},
			want: "processing",
		},
		{
			name:  "message",
			event: AgentEvent{Kind: "message", Data: json.RawMessage(`{"status":"ready"}`)},
		},
		{name: "invalid data", event: AgentEvent{Kind: "status", Data: json.RawMessage(`null`)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.Status(); got != tt.want {
				t.Errorf("status = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeAgentError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "agent detail",
			status: 404,
			body:   `{"detail":"Session not found"}`,
			want:   "agent returned 404: Session not found",
		},
		{
			name: "agent validation list", status: 422,
			body: `{"detail":[{"loc":["body","message"],"msg":"Field required"}]}`,
			want: "agent returned 422: message: Field required",
		},
		{
			name:   "platform message",
			status: 424,
			body:   `{"message":"Agent request could not be completed"}`,
			want:   "agent returned 424: Agent request could not be completed",
		},
		{
			name:   "plain text",
			status: 502,
			body:   "bad gateway\n",
			want:   "agent returned 502: bad gateway",
		},
		{name: "empty", status: 500, want: "agent returned 500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecodeAgentError(tt.status, []byte(tt.body)).Error(); got != tt.want {
				t.Errorf("error = %q, want %q", got, tt.want)
			}
		})
	}
}
