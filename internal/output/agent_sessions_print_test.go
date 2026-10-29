package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestPrintAgentSessionList(t *testing.T) {
	title := "Order 123"
	cursor := "c2"
	tests := []struct {
		name string
		list deployment.AgentSessionList
		want string
	}{
		{name: "empty", want: "No sessions found.\n"},
		{
			name: "rows and next page",
			list: deployment.AgentSessionList{
				Items: []deployment.AgentSession{
					{
						ID:         "s1",
						CustomerID: "guest",
						Title:      &title,
						Labels:     []string{"vip", "test", "eu", "beta"},
						CreatedAt:  "bad",
					},
					{ID: "s2", CustomerID: "c42", CreatedAt: "bad"},
				},
				TotalCount: 3,
				HasMore:    true,
				NextCursor: &cursor,
			},
			want: "ID   CUSTOMER   TITLE       LABELS                    CREATED\n" +
				"s1   guest      Order 123   vip, test, eu (+1 more)   bad\n" +
				"s2   c42                                              bad\n" +
				"\nMore results — next page: --cursor c2\n",
		},
		{
			name: "last page has no cursor line",
			list: deployment.AgentSessionList{
				Items: []deployment.AgentSession{
					{ID: "s1", CustomerID: "guest", CreatedAt: "bad"},
				},
				TotalCount: 1,
			},
			want: "ID   CUSTOMER   TITLE   LABELS   CREATED\n" +
				"s1   guest                       bad\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintAgentSessionList(&out, &tt.list); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Errorf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestPrintAgentEventList(t *testing.T) {
	tests := []struct {
		name   string
		events []deployment.AgentEvent
		want   string
	}{
		{name: "empty", want: "No events found.\n"},
		{
			name: "first line of each summary",
			events: []deployment.AgentEvent{
				{
					Offset:    0,
					ID:        "e0",
					Kind:      "message",
					Source:    "customer",
					CreatedAt: "bad",
					Data:      json.RawMessage(`{"message":"Hi\nthere"}`),
				},
				{
					Offset:    1,
					ID:        "e1",
					Kind:      "status",
					Source:    "agent",
					CreatedAt: "bad",
					Data:      json.RawMessage(`{"status":"ready"}`),
				},
			},
			want: "OFFSET   ID   KIND      SOURCE     CREATED   SUMMARY\n" +
				"0        e0   message   customer   bad       Hi\n" +
				"1        e1   status    agent      bad       ready\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintAgentEventList(&out, tt.events); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Errorf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestEventFollower(t *testing.T) {
	type update struct {
		event     deployment.AgentEvent
		streaming bool
	}
	partial := deployment.AgentEvent{
		Kind:   "message",
		Source: "agent",
		Data:   json.RawMessage(`{"message":"He","status":"streaming"}`),
		Raw:    json.RawMessage(`{"kind":"message","data":{"message":"He","status":"streaming"}}`),
	}
	complete := deployment.AgentEvent{
		Kind:   "message",
		Source: "agent",
		Data:   json.RawMessage(`{"message":"Hello","status":"completed"}`),
		Raw: json.RawMessage(
			`{"kind":"message","data":{"message":"Hello","status":"completed"}}`,
		),
	}
	status := deployment.AgentEvent{
		Kind:   "status",
		Source: "agent",
		Data:   json.RawMessage(`{"status":"ready"}`),
		Raw:    json.RawMessage(`{"kind":"status","data":{"status":"ready"}}`),
	}
	updates := []update{{partial, true}, {complete, false}, {status, false}}
	tests := []struct {
		name   string
		asJSON bool
		asYAML bool
		want   string
	}{
		{
			name:   "JSON writes each complete event as one line",
			asJSON: true,
			want: `{"kind":"message","data":{"message":"Hello","status":"completed"}}` + "\n" +
				`{"kind":"status","data":{"status":"ready"}}` + "\n",
		},
		{
			name:   "YAML writes each complete event as one document",
			asYAML: true,
			want: "---\ndata:\n    message: Hello\n    status: completed\nkind: message\n" +
				"---\ndata:\n    status: ready\nkind: status\n",
		},
		{
			name: "plain lines when not a terminal",
			want: "agent: Hello\n(ready)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			print := EventFollower(&out, tt.asJSON, tt.asYAML)
			for _, u := range updates {
				if err := print(u.event, u.streaming); err != nil {
					t.Fatal(err)
				}
			}
			if out.String() != tt.want {
				t.Errorf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}
