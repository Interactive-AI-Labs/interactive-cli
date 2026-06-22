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

func TestSessionSummary_Empty(t *testing.T) {
	m := SessionSummary("s_x", nil)
	if m.ID != "s_x" || m.TurnCount != 0 || len(m.Turns) != 0 {
		t.Fatalf("empty session = %+v", m)
	}
}
