package summary

import "testing"

func TestTraceDiff(t *testing.T) {
	a := &TraceSummaryModel{
		Name: "agent-chat",
		Cost: ptr(0.09),
		Iterations: []Iteration{
			{
				Number:   1,
				Routines: []string{"bonus-chat"},
				Journey:  []JourneyStep{{Routine: "bonus-chat", Step: "first_tool"}},
				Tools:    []ToolCall{{Name: "authenticate_customer"}},
			},
			{Number: 2, Tools: []ToolCall{{Name: "get_user_transactions"}}},
			{
				Number: 3,
				Journey: []JourneyStep{
					{Routine: "bonus-chat", Step: "elig_inquiry_no_deposit_eligible"},
				},
			},
		},
		Reply: "Vous êtes éligible",
	}
	b := &TraceSummaryModel{
		Name: "agent-chat",
		Cost: ptr(0.20),
		Iterations: []Iteration{
			{
				Number:   1,
				Routines: []string{"bonus-chat"},
				Journey:  []JourneyStep{{Routine: "bonus-chat", Step: "first_tool"}},
				Tools:    []ToolCall{{Name: "authenticate_customer"}},
			},
			{
				Number:   2,
				Routines: []string{"player-login-issue-chat"},
				Tools: []ToolCall{
					{Name: "get_user_transactions"},
					{Name: "get_filtered_action_log"},
				},
			},
			{
				Number:  3,
				Journey: []JourneyStep{{Routine: "bonus-chat", Step: "elig_inquiry_not_eligible"}},
				Tools:   []ToolCall{{Name: "initiate_human_handoff"}},
			},
		},
		Reply: "transfer to human",
	}

	want := `{
		"a":{"id":"A","name":"agent-chat","iterations":3,"cost":0.09,"reply":"Vous êtes éligible"},
		"b":{"id":"B","name":"agent-chat","iterations":3,"cost":0.2,"reply":"transfer to human"},
		"routines":{"both":["bonus-chat"],"b_only":["player-login-issue-chat"]},
		"tools":{
			"both":["authenticate_customer","get_user_transactions"],
			"b_only":["get_filtered_action_log","initiate_human_handoff"]
		},
		"journey":[
			{"iteration":1,"a":["bonus-chat/first_tool"],"b":["bonus-chat/first_tool"]},
			{"iteration":3,
			 "a":["bonus-chat/elig_inquiry_no_deposit_eligible"],
			 "b":["bonus-chat/elig_inquiry_not_eligible"],
			 "diverged":true}
		]
	}`
	assertJSON(t, TraceDiff("A", a, "B", b), want)
}

func ptr(v float64) *float64 { return &v }
