package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

func TestPrintTraceDiff(t *testing.T) {
	cases := []struct {
		name   string
		model  *summary.TraceDiffModel
		want   []string
		absent []string
	}{
		{
			name: "full diff with divergence and replies",
			model: &summary.TraceDiffModel{
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
			},
			want: []string{
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
			},
		},
		{
			name: "no journey steps omits journey section",
			model: &summary.TraceDiffModel{
				A:        summary.DiffSide{ID: "ta", Iterations: 1, Reply: "hi"},
				B:        summary.DiffSide{ID: "tb", Iterations: 1, Reply: "hello"},
				Routines: summary.SetDiff{Both: []string{"bonus-chat"}},
			},
			want:   []string{"ta", "tb", "both: bonus-chat", "Reply"},
			absent: []string{"Journey", "◀ diverges"},
		},
		{
			name: "empty replies omit reply section",
			model: &summary.TraceDiffModel{
				A:     summary.DiffSide{ID: "ta", Iterations: 1},
				B:     summary.DiffSide{ID: "tb", Iterations: 1},
				Tools: summary.SetDiff{AOnly: []string{"authenticate_customer"}},
			},
			want:   []string{"ta", "tb", "A only: authenticate_customer"},
			absent: []string{"Reply", "Journey"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintTraceDiff(&buf, tc.model); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("output missing %q\n---\n%s", want, got)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(got, absent) {
					t.Fatalf("output should not contain %q\n---\n%s", absent, got)
				}
			}
		})
	}
}
