package summary

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestSessionSummary(t *testing.T) {
	cases := []struct {
		name      string
		sessionID string
		traces    []clients.TraceInfo
		want      string
	}{
		{
			name:      "transcript with tools and a journey",
			sessionID: "s_abc",
			traces: []clients.TraceInfo{
				{
					ID: "t1", Name: "turn", Timestamp: "2026-06-22T14:30:00Z",
					Tags:   []string{"agent:driveaway-agent", "tool:check_availability"},
					Input:  json.RawMessage(`"I want to rent a car next weekend"`),
					Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
				},
				{
					ID:        "t2",
					Name:      "turn",
					Timestamp: "2026-06-22T14:32:00Z",
					Tags: []string{
						"agent:driveaway-agent",
						"tool:create_booking",
						"routine:rental",
					},
					Input:  json.RawMessage(`"Downtown"`),
					Output: json.RawMessage(`"[\"Booked! Confirmation #1234\"]"`),
				},
			},
			want: `{
				"id":"s_abc","agent":"driveaway-agent","turn_count":2,"duration":"2m0s",
				"turns":[
					{"number":1,"customer":"I want to rent a car next weekend","agent":"Great! We have 3 cars available...","tools":["check_availability"]},
					{"number":2,"customer":"Downtown","agent":"Booked! Confirmation #1234","tools":["create_booking"],"journeys":["rental"]}
				]
			}`,
		},
		{
			name:      "multiple agents surface deduped in order",
			sessionID: "s_multi",
			traces: []clients.TraceInfo{
				{ID: "t1", Timestamp: "2026-06-22T14:30:00Z", Tags: []string{"agent:agent-chat"}},
				{
					ID:        "t2",
					Timestamp: "2026-06-22T14:31:00Z",
					Tags:      []string{"agent:agent-chat-dev"},
				},
				{ID: "t3", Timestamp: "2026-06-22T14:32:00Z", Tags: []string{"agent:agent-chat"}},
			},
			want: `{
				"id":"s_multi","agent":"agent-chat, agent-chat-dev","turn_count":3,"duration":"2m0s",
				"turns":[{"number":1},{"number":2},{"number":3}]
			}`,
		},
		{
			name:      "empty session",
			sessionID: "s_x",
			traces:    nil,
			want:      `{"id":"s_x","turn_count":0}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertJSON(t, SessionSummary(tc.sessionID, tc.traces), tc.want)
		})
	}
}

func TestSessionSummary_Cost(t *testing.T) {
	c := func(v float64) *float64 { return &v }
	cases := []struct {
		name     string
		traces   []clients.TraceInfo
		wantNil  bool
		wantCost float64 // checked with tolerance; ignored when wantNil
	}{
		{
			name: "sums non-nil costs, ignores gaps",
			traces: []clients.TraceInfo{
				{ID: "t1", Timestamp: "2026-06-22T14:30:00Z", TotalCost: c(0.05)},
				{ID: "t2", Timestamp: "2026-06-22T14:31:00Z"}, // nil cost
				{ID: "t3", Timestamp: "2026-06-22T14:32:00Z", TotalCost: c(0.03)},
			},
			wantCost: 0.08,
		},
		{
			name:    "no costs anywhere yields nil",
			traces:  []clients.TraceInfo{{ID: "t1", Timestamp: "2026-06-22T14:30:00Z"}},
			wantNil: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := SessionSummary("s", tc.traces)
			if tc.wantNil {
				if m.Cost != nil {
					t.Fatalf("Cost = %v, want nil", m.Cost)
				}
				return
			}
			if m.Cost == nil || *m.Cost < tc.wantCost-0.001 || *m.Cost > tc.wantCost+0.001 {
				t.Fatalf("Cost = %v, want ~%v", m.Cost, tc.wantCost)
			}
		})
	}
}
