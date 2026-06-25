package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

func f64(v float64) *float64 { return &v }

func TestPrintTraceSummary(t *testing.T) {
	cases := []struct {
		name  string
		model *summary.TraceSummaryModel
		want  []string
	}{
		{
			name: "iterations, conditions, tools and reply",
			model: &summary.TraceSummaryModel{
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
						Tools: []summary.ToolCall{{
							Name:   "check_availability",
							Args:   json.RawMessage(`{"dates":"next weekend"}`),
							Result: json.RawMessage(`{"count":3}`),
						}},
					},
					{
						Number: 2,
						Conditions: []summary.Condition{
							{Text: "Pickup location not yet provided", Score: 8},
						},
					},
				},
				Reply: "Great! We have 3 cars available...",
			},
			want: []string{
				"driveaway-agent",
				"2 iterations",
				"Customer:\n  I want to rent a car for next weekend",
				"Iteration 1",
				"Customer asks to rent a vehicle (9)",
				"    → check_availability(dates=\"next weekend\")\n      Result:\n        {\"count\":3}",
				"Iteration 2",
				"(no tools called)",
				"Agent:\n  Great! We have 3 cars available...",
			},
		},
		{
			name: "journey decision path with routines and rationale",
			model: &summary.TraceSummaryModel{
				Name:  "agent-chat",
				Input: "first bet refund?",
				Iterations: []summary.Iteration{{
					Number:   1,
					Routines: []string{"bonus-chat", "player-login-issue-chat"},
					Journey: []summary.JourneyStep{{
						Routine:   "bonus-chat",
						Step:      "elig_inquiry_not_eligible",
						Condition: "MainMoneyBet OR decommission tag → escalate",
					}},
					Decisions: []string{"TAGS show decommission, condition 4 fits"},
				}},
				Reply: "transfer",
			},
			want: []string{
				"Routines: bonus-chat, player-login-issue-chat",
				"Journey:",
				"bonus-chat ▸ elig_inquiry_not_eligible",
				"MainMoneyBet OR decommission tag → escalate",
				"Why: TAGS show decommission, condition 4 fits",
			},
		},
		{
			name: "titled knowledge base",
			model: &summary.TraceSummaryModel{
				Name:  "agent-chat",
				Input: "hi",
				KB: &summary.KBRetrieval{
					Docs:  []string{"Closing my account", "Why suspended"},
					Count: 2,
				},
				Iterations: []summary.Iteration{{Number: 1}},
				Reply:      "hello",
			},
			want: []string{
				`Knowledge base: 2 docs retrieved — "Closing my account", "Why suspended"`,
			},
		},
		{
			name: "untitled knowledge base reports count only",
			model: &summary.TraceSummaryModel{
				Name:       "agent",
				Input:      "hi",
				KB:         &summary.KBRetrieval{Count: 3},
				Iterations: []summary.Iteration{{Number: 1}},
			},
			want: []string{"Knowledge base: 3 docs retrieved\n"},
		},
		{
			name: "trace error and tool error",
			model: &summary.TraceSummaryModel{
				Name:  "agent",
				Level: "ERROR",
				Input: "hi",
				Iterations: []summary.Iteration{
					{
						Number: 1,
						Tools: []summary.ToolCall{
							{Name: "create_booking", Errored: true, ErrMsg: "upstream 500"},
						},
					},
				},
				Reply:  "sorry",
				Errors: []string{"create_booking: upstream 500"},
			},
			want: []string{
				"ERROR",
				"    → create_booking()",
				"      ERROR:\n        upstream 500",
				"Errors:",
			},
		},
		{
			name: "multiline reply keeps its structure as an indented block",
			model: &summary.TraceSummaryModel{
				Name:       "agent",
				Input:      "hi",
				Iterations: []summary.Iteration{{Number: 1}},
				Reply:      "Bonjour,\n\nMerci de nous avoir contactés.",
			},
			want: []string{"Agent:\n  Bonjour,", "\n  Merci de nous avoir contactés."},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintTraceSummary(&buf, tc.model); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("output missing %q\n---\n%s", want, got)
				}
			}
		})
	}
}
