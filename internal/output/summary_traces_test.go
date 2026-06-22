package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

func f64(v float64) *float64 { return &v }

func TestPrintTraceSummary(t *testing.T) {
	m := &summary.TraceSummaryModel{
		Name:      "driveaway-agent",
		Timestamp: "2026-06-22T14:32:01Z",
		LatencyMs: f64(4200),
		Cost:      f64(0.012),
		Level:     "DEFAULT",
		Input:     "I want to rent a car for next weekend",
		Iterations: []summary.Iteration{
			{
				Number: 1,
				Conditions: []summary.Condition{
					{Text: "Customer asks to rent a vehicle", Score: 9},
				},
				Tools: []summary.ToolCall{
					{
						Name:   "check_availability",
						Args:   `dates="next weekend"`,
						Result: `{"count":3}`,
					},
				},
			},
			{
				Number: 2,
				Conditions: []summary.Condition{
					{Text: "Pickup location not yet provided", Score: 8},
				},
			},
		},
		Reply: "Great! We have 3 cars available...",
	}

	var buf bytes.Buffer
	if err := PrintTraceSummary(&buf, m); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"driveaway-agent",
		"2 iterations",
		"Customer: I want to rent a car for next weekend",
		"Iteration 1",
		"Customer asks to rent a vehicle (9)",
		"check_availability(dates=\"next weekend\") → {\"count\":3}",
		"Iteration 2",
		"(no tools called)",
		"Agent: Great! We have 3 cars available...",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n---\n%s", want, got)
		}
	}
}

func TestPrintTraceSummary_KnowledgeBase(t *testing.T) {
	m := &summary.TraceSummaryModel{
		Name:  "agent-chat",
		Input: "hi",
		KB: &summary.KBRetrieval{
			Docs:  []string{"Closing my account", "Why suspended"},
			Count: 2,
		},
		Iterations: []summary.Iteration{{Number: 1}},
		Reply:      "hello",
	}
	var buf bytes.Buffer
	if err := PrintTraceSummary(&buf, m); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		`Knowledge base: 2 docs retrieved — "Closing my account", "Why suspended"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n---\n%s", want, got)
		}
	}
}

func TestPrintTraceSummary_KnowledgeBaseUntitled(t *testing.T) {
	m := &summary.TraceSummaryModel{
		Name:       "agent",
		Input:      "hi",
		KB:         &summary.KBRetrieval{Count: 3},
		Iterations: []summary.Iteration{{Number: 1}},
	}
	var buf bytes.Buffer
	if err := PrintTraceSummary(&buf, m); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Knowledge base: 3 docs retrieved\n") {
		t.Fatalf("untitled KB line wrong:\n%s", buf.String())
	}
}

func TestPrintTraceSummary_ErrorAndToolError(t *testing.T) {
	m := &summary.TraceSummaryModel{
		Name:  "agent",
		Level: "ERROR",
		Input: "hi",
		Iterations: []summary.Iteration{
			{
				Number: 1,
				Tools: []summary.ToolCall{
					{Name: "create_booking", Args: "", Errored: true, ErrMsg: "upstream 500"},
				},
			},
		},
		Reply:  "sorry",
		Errors: []string{"create_booking: upstream 500"},
	}
	var buf bytes.Buffer
	if err := PrintTraceSummary(&buf, m); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"ERROR", "create_booking(", "ERROR: upstream 500", "Errors:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n---\n%s", want, got)
		}
	}
}
