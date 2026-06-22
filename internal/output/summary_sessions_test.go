package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

func TestPrintSessionSummary(t *testing.T) {
	m := &summary.SessionSummaryModel{
		ID: "s_abc", Agent: "driveaway-agent", TurnCount: 2,
		Duration: "2m0s", Cost: f64(0.08),
		Turns: []summary.Turn{
			{
				Number: 1, Customer: "I want to rent a car next weekend",
				Agent: "Great! We have 3 cars available...", Tools: []string{"check_availability"},
			},
			{
				Number: 2, Customer: "Downtown", Agent: "Booked! Confirmation #1234",
				Tools: []string{"create_booking"}, Journeys: []string{"rental"},
			},
		},
	}
	var buf bytes.Buffer
	if err := PrintSessionSummary(&buf, m); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"Session s_abc · driveaway-agent · 2 turns · 2m0s",
		"Turn 1",
		"Customer: I want to rent a car next weekend",
		"[tools: check_availability]",
		"Agent: Great! We have 3 cars available...",
		"Turn 2",
		"[tools: create_booking] [journey: rental]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n---\n%s", want, got)
		}
	}
}

func TestPrintSessionSummary_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintSessionSummary(&buf, &summary.SessionSummaryModel{ID: "s_x"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No turns found.") {
		t.Fatalf("expected empty-state message, got %q", buf.String())
	}
}
