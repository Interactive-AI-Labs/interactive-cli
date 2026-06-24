package summary

import "testing"

func TestTraceDiff(t *testing.T) {
	cases := []struct {
		name string
		a, b *TraceSummaryModel
		want string
	}{
		{
			name: "diverging journeys, tools, and routines",
			a: &TraceSummaryModel{
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
			},
			b: &TraceSummaryModel{
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
						Number: 3,
						Journey: []JourneyStep{
							{Routine: "bonus-chat", Step: "elig_inquiry_not_eligible"},
						},
						Tools: []ToolCall{{Name: "initiate_human_handoff"}},
					},
				},
				Reply: "transfer to human",
			},
			want: `{
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
			}`,
		},
		{
			name: "identical turns report no divergence",
			a: &TraceSummaryModel{
				Name: "agent-chat",
				Iterations: []Iteration{{
					Number:   1,
					Routines: []string{"bonus-chat"},
					Journey:  []JourneyStep{{Routine: "bonus-chat", Step: "first_tool"}},
					Tools:    []ToolCall{{Name: "authenticate_customer"}},
				}},
			},
			b: &TraceSummaryModel{
				Name: "agent-chat",
				Iterations: []Iteration{{
					Number:   1,
					Routines: []string{"bonus-chat"},
					Journey:  []JourneyStep{{Routine: "bonus-chat", Step: "first_tool"}},
					Tools:    []ToolCall{{Name: "authenticate_customer"}},
				}},
			},
			want: `{
				"a":{"id":"A","name":"agent-chat","iterations":1},
				"b":{"id":"B","name":"agent-chat","iterations":1},
				"routines":{"both":["bonus-chat"]},
				"tools":{"both":["authenticate_customer"]},
				"journey":[{"iteration":1,"a":["bonus-chat/first_tool"],"b":["bonus-chat/first_tool"]}]
			}`,
		},
		{
			name: "routines and tools only on A",
			a: &TraceSummaryModel{
				Name: "agent-chat",
				Iterations: []Iteration{{
					Number:   1,
					Routines: []string{"bonus-chat"},
					Tools:    []ToolCall{{Name: "authenticate_customer"}},
				}},
			},
			b: &TraceSummaryModel{Name: "agent-chat", Iterations: []Iteration{{Number: 1}}},
			want: `{
				"a":{"id":"A","name":"agent-chat","iterations":1},
				"b":{"id":"B","name":"agent-chat","iterations":1},
				"routines":{"a_only":["bonus-chat"]},
				"tools":{"a_only":["authenticate_customer"]}
			}`,
		},
		{
			name: "no journey steps yields no journey section",
			a:    &TraceSummaryModel{Name: "agent-chat", Iterations: []Iteration{{Number: 1}}},
			b:    &TraceSummaryModel{Name: "agent-chat", Iterations: []Iteration{{Number: 1}}},
			want: `{
				"a":{"id":"A","name":"agent-chat","iterations":1},
				"b":{"id":"B","name":"agent-chat","iterations":1},
				"routines":{},"tools":{}
			}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertJSON(t, TraceDiff("A", tc.a, "B", tc.b), tc.want)
		})
	}
}

func TestEqualStringSets(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"both empty", nil, nil, true},
		{"same members", []string{"x", "y"}, []string{"y", "x"}, true},
		{"different members", []string{"x"}, []string{"y"}, false},
		{"different sizes", []string{"x"}, []string{"x", "y"}, false},
		{"duplicates collapse to same set", []string{"x", "x"}, []string{"x"}, true},
		{"duplicate masks a distinct member", []string{"x", "x"}, []string{"x", "y"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := equalStringSets(tc.a, tc.b); got != tc.want {
				t.Fatalf("equalStringSets(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func ptr(v float64) *float64 { return &v }
