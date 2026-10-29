package output

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestEventBlock(t *testing.T) {
	labels := map[string]string{"customer": "you", "agent": "hotel"}
	tests := []struct {
		name   string
		event  deployment.AgentEvent
		labels map[string]string
		width  int
		want   string
	}{
		{
			name: "agent message labelled and wrapped",
			event: deployment.AgentEvent{
				Kind:   "message",
				Source: "agent",
				Data:   json.RawMessage(`{"message":"Welcome to the hotel"}`),
			},
			labels: labels,
			width:  13,
			want:   "\n" + colorBrand + "●" + colorReset + " " + colorBold + "hotel" + colorReset + "\n  Welcome to\n  the hotel\n",
		},
		{
			name: "customer message without a known width",
			event: deployment.AgentEvent{
				Kind:   "message",
				Source: "customer",
				Data:   json.RawMessage(`{"message":"Hi"}`),
			},
			labels: labels,
			want:   "\n" + colorBlue + "●" + colorReset + " " + colorBold + "you" + colorReset + "\n  Hi\n",
		},
		{
			name: "unlabelled source keeps its name",
			event: deployment.AgentEvent{
				Kind:   "message",
				Source: "human",
				Data:   json.RawMessage(`{"message":"One\ntwo"}`),
			},
			want: "\n" + colorBlue + "●" + colorReset + " " + colorBold + "human" + colorReset + "\n  One\n  two\n",
		},
		{
			name: "status is a dim line",
			event: deployment.AgentEvent{
				Kind:   "status",
				Source: "agent",
				Data:   json.RawMessage(`{"status":"ready"}`),
			},
			want: "  " + colorDim + "· ready" + colorReset + "\n",
		},
		{
			name: "other events truncated to the width",
			event: deployment.AgentEvent{
				Kind:   "custom",
				Source: "system",
				Data:   json.RawMessage(`{"key":"value"}`),
			},
			width: 20,
			want:  "  " + colorDim + "· custom syste..." + colorReset + "\n",
		},
		{
			name: "too narrow to truncate",
			event: deployment.AgentEvent{
				Kind:   "custom",
				Source: "system",
				Data:   json.RawMessage(`{"key":"value"}`),
			},
			width: 5,
			want:  "  " + colorDim + "· custom system: {\"key\":\"value\"}" + colorReset + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EventBlock(tt.event, tt.labels, tt.width); got != tt.want {
				t.Errorf("block = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChatBanner(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		sessionID string
		want      string
	}{
		{
			name:      "mark beside the wordmark and session",
			agentName: "hotel",
			sessionID: "s1",
			want: colorBrand + "  ██▄▄██  " + colorReset + "  " + colorBold + "interactive AI" + colorReset + "\n" +
				colorBrand + "▄▄  ██  ▄▄" + colorReset + "\n" +
				colorBrand + "██  ██  ██" + colorReset + "  Chatting with " + colorBold + "hotel" + colorReset + "\n" +
				colorBrand + "▀▀  ██  ▀▀" + colorReset + "  " + colorDim + "Session s1 · resume with --session-id s1" + colorReset + "\n" +
				colorBrand + "  ██▀▀██  " + colorReset + "  " + colorDim + "Ctrl+C or Ctrl+D to leave" + colorReset + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChatBanner(tt.agentName, tt.sessionID); got != tt.want {
				t.Errorf("banner = %q, want %q", got, tt.want)
			}
		})
	}
}
