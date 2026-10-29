package output

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestEventLine(t *testing.T) {
	tests := []struct {
		name   string
		event  deployment.AgentEvent
		labels map[string]string
		want   string
	}{
		{
			name: "customer message",
			event: deployment.AgentEvent{
				Kind:   "message",
				Source: "customer",
				Data:   json.RawMessage(`{"message":"Hello"}`),
			},
			want: "customer: Hello",
		},
		{
			name: "message sender labelled",
			event: deployment.AgentEvent{
				Kind:   "message",
				Source: "agent",
				Data:   json.RawMessage(`{"message":"Hi"}`),
			},
			labels: map[string]string{"agent": "hotel"},
			want:   "hotel: Hi",
		},
		{
			name: "status",
			event: deployment.AgentEvent{
				Kind:   "status",
				Source: "agent",
				Data:   json.RawMessage(`{"status":"typing","data":{}}`),
			},
			want: "(typing)",
		},
		{
			name: "tool event shows compact data",
			event: deployment.AgentEvent{
				Kind:   "tool",
				Source: "agent",
				Data:   json.RawMessage("{\"toolCalls\": [ {\"toolId\": \"lookup\"} ]}"),
			},
			want: `tool agent: {"toolCalls":[{"toolId":"lookup"}]}`,
		},
		{
			name: "invalid data shown as is",
			event: deployment.AgentEvent{
				Kind:   "custom",
				Source: "system",
				Data:   json.RawMessage(`{broken`),
			},
			want: "custom system: {broken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EventLine(tt.event, tt.labels); got != tt.want {
				t.Errorf("line = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{name: "no width", text: "a long line", width: 0, want: "a long line"},
		{name: "fits", text: "hello world", width: 11, want: "hello world"},
		{name: "breaks at space", text: "hello wide world", width: 10, want: "hello wide\nworld"},
		{
			name:  "breaks before a word that does not fit",
			text:  "hello world",
			width: 8,
			want:  "hello\nworld",
		},
		{
			name:  "cuts a word longer than the width",
			text:  "abcdefghij",
			width: 4,
			want:  "abcd\nefgh\nij",
		},
		{
			name:  "keeps existing newlines",
			text:  "one\ntwo three",
			width: 5,
			want:  "one\ntwo\nthree",
		},
		{name: "counts runes", text: "ññññ ññ", width: 4, want: "ññññ\nññ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WrapText(tt.text, tt.width); got != tt.want {
				t.Errorf("wrapped = %q, want %q", got, tt.want)
			}
		})
	}
}
