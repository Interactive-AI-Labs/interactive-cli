package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

func TestPrintTraceDiff(t *testing.T) {
	m := &summary.TraceDiffModel{
		A: summary.DiffSide{
			ID: "ta", Name: "agent-chat", Iterations: 3, Cost: f64(0.089),
			Reply: "Vous êtes éligible",
		},
		B: summary.DiffSide{
			ID: "tb", Name: "agent-chat", Iterations: 5, Cost: f64(0.203),
			Reply: "transfer to human",
		},
		Routines: summary.SetDiff{
			Both:  []string{"bonus-chat"},
			BOnly: []string{"player-login-issue-chat"},
		},
		Tools: summary.SetDiff{
			Both:  []string{"authenticate_customer", "get_user_transactions"},
			BOnly: []string{"get_filtered_action_log", "initiate_human_handoff"},
		},
		Journey: []summary.JourneyDiff{
			{
				Iteration: 1,
				A:         []string{"bonus-chat/first_tool"},
				B:         []string{"bonus-chat/first_tool"},
			},
			{
				Iteration: 3,
				A:         []string{"bonus-chat/elig_inquiry_no_deposit_eligible"},
				B:         []string{"bonus-chat/elig_inquiry_not_eligible"},
				Diverged:  true,
			},
		},
	}

	var buf bytes.Buffer
	if err := PrintTraceDiff(&buf, m); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"ta", "tb",
		"agent-chat · 3 iters",
		"agent-chat · 5 iters",
		"both: bonus-chat",
		"B only: player-login-issue-chat",
		"both: authenticate_customer, get_user_transactions",
		"B only: get_filtered_action_log, initiate_human_handoff",
		"iter 3",
		"bonus-chat/elig_inquiry_not_eligible",
		"◀ diverges",
		"Vous êtes éligible",
		"transfer to human",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n---\n%s", want, got)
		}
	}
}
