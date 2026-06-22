package summary

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestSessionSummary(t *testing.T) {
	traces := []clients.TraceInfo{
		{
			ID: "t1", Name: "turn", Timestamp: "2026-06-22T14:30:00Z",
			Tags:   []string{"agent:driveaway-agent", "tool:check_availability"},
			Input:  json.RawMessage(`"I want to rent a car next weekend"`),
			Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
		},
		{
			ID: "t2", Name: "turn", Timestamp: "2026-06-22T14:32:00Z",
			Tags:   []string{"agent:driveaway-agent", "tool:create_booking", "routine:rental"},
			Input:  json.RawMessage(`"Downtown"`),
			Output: json.RawMessage(`"[\"Booked! Confirmation #1234\"]"`),
		},
	}

	m := SessionSummary("s_abc", traces)

	if m.ID != "s_abc" || m.Agent != "driveaway-agent" || m.TurnCount != 2 {
		t.Fatalf("header = %+v", m)
	}
	if m.Duration != "2m0s" {
		t.Fatalf("Duration = %q", m.Duration)
	}
	if len(m.Turns) != 2 {
		t.Fatalf("turns = %d", len(m.Turns))
	}
	if m.Turns[0].Customer != "I want to rent a car next weekend" ||
		m.Turns[0].Agent != "Great! We have 3 cars available..." {
		t.Fatalf("turn1 = %+v", m.Turns[0])
	}
	if len(m.Turns[0].Tools) != 1 || m.Turns[0].Tools[0] != "check_availability" {
		t.Fatalf("turn1 tools = %+v", m.Turns[0].Tools)
	}
	if len(m.Turns[1].Journeys) != 1 || m.Turns[1].Journeys[0] != "rental" {
		t.Fatalf("turn2 journeys = %+v", m.Turns[1].Journeys)
	}
}

func TestSessionSummary_MultipleAgents(t *testing.T) {
	// A shadow/dev deployment logging alongside production: both agents surface.
	traces := []clients.TraceInfo{
		{ID: "t1", Timestamp: "2026-06-22T14:30:00Z", Tags: []string{"agent:agent-chat"}},
		{ID: "t2", Timestamp: "2026-06-22T14:31:00Z", Tags: []string{"agent:agent-chat-dev"}},
		{ID: "t3", Timestamp: "2026-06-22T14:32:00Z", Tags: []string{"agent:agent-chat"}},
	}
	m := SessionSummary("s_multi", traces)
	if m.Agent != "agent-chat, agent-chat-dev" {
		t.Fatalf("Agent = %q, want both agents deduped in order", m.Agent)
	}
}

func TestSessionSummary_Empty(t *testing.T) {
	m := SessionSummary("s_x", nil)
	if m.ID != "s_x" || m.TurnCount != 0 || len(m.Turns) != 0 {
		t.Fatalf("empty session = %+v", m)
	}
}

func TestSessionSummary_Cost(t *testing.T) {
	c := func(v float64) *float64 { return &v }
	traces := []clients.TraceInfo{
		{ID: "t1", Timestamp: "2026-06-22T14:30:00Z", TotalCost: c(0.05)},
		{ID: "t2", Timestamp: "2026-06-22T14:31:00Z"}, // nil cost
		{ID: "t3", Timestamp: "2026-06-22T14:32:00Z", TotalCost: c(0.03)},
	}
	m := SessionSummary("s_cost", traces)
	if m.Cost == nil || *m.Cost < 0.079 || *m.Cost > 0.081 {
		t.Fatalf("Cost = %v, want ~0.08", m.Cost)
	}
	// no costs anywhere -> nil
	m2 := SessionSummary(
		"s_nocost",
		[]clients.TraceInfo{{ID: "t1", Timestamp: "2026-06-22T14:30:00Z"}},
	)
	if m2.Cost != nil {
		t.Fatalf("Cost = %v, want nil", m2.Cost)
	}
}
