package summary

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

func obs(id, parent, typ, name, level, status string, in, out string) api.ObservationInfo {
	o := api.ObservationInfo{
		ID: id, ParentObservationID: parent, Type: typ, Name: name,
		Level: level, StatusMessage: status,
	}
	if in != "" {
		o.Input = json.RawMessage(in)
	}
	if out != "" {
		o.Output = json.RawMessage(out)
	}
	return o
}

func obsAt(id, parent, name, start string) api.ObservationInfo {
	o := obs(id, parent, "CHAIN", name, "", "", "", "")
	o.StartTime = start
	return o
}

// assertJSON compares got to wantJSON structurally, ignoring key order and whitespace.
func assertJSON(t *testing.T, got any, wantJSON string) {
	t.Helper()
	gotBytes, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	var gotV, wantV any
	if err := json.Unmarshal(gotBytes, &gotV); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(wantJSON), &wantV); err != nil {
		t.Fatalf("bad wantJSON %q: %v", wantJSON, err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Fatalf("model mismatch\n got: %s\nwant: %s", gotBytes, wantJSON)
	}
}

func TestTraceSummary(t *testing.T) {
	cases := []struct {
		name  string
		trace *api.TraceDetail
		obs   []api.ObservationInfo
		want  string
	}{
		{
			name: "two iterations with conditions and a tool",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name:   "driveaway-agent",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"I want to rent a car for next weekend"`),
				Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
			}},
			obs: []api.ObservationInfo{
				obs("it1", "root", "chain", "preparation_iteration_1", "", "", "", ""),
				obs(
					"mg1",
					"it1",
					"chain",
					"match_guidelines",
					"",
					"",
					"",
					`{"matches":[{"condition":"Customer asks to rent a vehicle","score":9},{"condition":"No booking in progress","score":7}]}`,
				),
				obs("ex1", "it1", "tool", "execute_tool_calls", "", "", "", ""),
				obs("t1", "ex1", "tool", "check_availability", "", "",
					`"{\"dates\":\"next weekend\"}"`, `{"count":3}`),
				obs("it2", "root", "chain", "preparation_iteration_2", "", "", "", ""),
				obs("mg2", "it2", "chain", "match_guidelines", "", "",
					"", `{"matches":[{"condition":"Pickup location not yet provided","score":8}]}`),
			},
			want: `{
				"name":"driveaway-agent","level":"DEFAULT",
				"input":"I want to rent a car for next weekend",
				"iterations":[
					{"number":1,
					 "conditions":[{"text":"Customer asks to rent a vehicle","score":9},{"text":"No booking in progress","score":7}],
					 "tools":[{"name":"check_availability","args":{"dates":"next weekend"},"result":{"count":3}}]},
					{"number":2,"conditions":[{"text":"Pickup location not yet provided","score":8}]}
				],
				"reply":"Great! We have 3 cars available..."
			}`,
		},
		{
			name: "journey path, routine activation, policy, and decision rationale",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name:   "agent-chat",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"first bet refund?"`),
				Output: json.RawMessage(`"[\"transfer\"]"`),
			}},
			obs: []api.ObservationInfo{
				obs("it1", "root", "chain", "preparation_iteration_1", "", "", "", ""),
				obs("mg1", "it1", "chain", "match_guidelines", "", "", "", `{"matches":[
					{"type":"routine","routine_id":"bonus-chat","condition":"big routine cond","score":10},
					{"type":"routine","routine_id":"bonus-chat","condition":"big routine cond","score":10},
					{"type":"routine_node","routine_id":"bonus-chat","step_id":"first_tool","condition":"","score":10},
					{"type":"routine_node","routine_id":"bonus-chat","step_id":"elig_inquiry_not_eligible","condition":"MainMoneyBet  OR decommission","score":10},
					{"type":"routine_node","routine_id":"bonus-chat","step_id":"elig_inquiry_not_eligible","condition":"MainMoneyBet OR decommission","score":10},
					{"type":"policy","id":"handoff","condition":"Always applies.","score":10}
				]}`),
				// Winning next-step decision plus a dropped incomplete (applied_condition_id "0") one.
				obs("ns0", "mg1", "generation", "next-step", "", "", "",
					`{"applied_condition_id":"0","next_step_rationale":"step incomplete"}`),
				obs(
					"ns1",
					"mg1",
					"generation",
					"next-step",
					"",
					"",
					"",
					`{"applied_condition_id":"4","next_step_rationale":"TAGS show decommission, condition 4 fits"}`,
				),
			},
			want: `{
				"name":"agent-chat","level":"DEFAULT","input":"first bet refund?","reply":"transfer",
				"iterations":[{"number":1,
					"routines":["bonus-chat"],
					"journey":[
						{"routine":"bonus-chat","step":"first_tool"},
						{"routine":"bonus-chat","step":"elig_inquiry_not_eligible","condition":"MainMoneyBet OR decommission"}
					],
					"conditions":[{"text":"Always applies.","score":10}],
					"decisions":["TAGS show decommission, condition 4 fits"]
				}]
			}`,
		},
		{
			name: "titled knowledge-base retrieval at the root",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name:   "agent-chat",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"hi"`),
				Output: json.RawMessage(`"[\"hello\"]"`),
			}},
			obs: []api.ObservationInfo{
				obs(
					"kb",
					"process",
					"span",
					"retriever:knowledge_base",
					"",
					"",
					`{"customer_messages":["hi"],"customer_id":"1"}`,
					`{"has_results":true,"article_count":2,"articles":[{"name":"Closing my account"},{"name":"Why was my account suspended"}]}`,
				),
				obs("it1", "process", "chain", "preparation_iteration_1", "", "", "", ""),
				// Per-iteration untitled vector search: its context-blob query must not leak.
				obs("fs1", "it1", "span", "find_similar_documents", "", "",
					`{"query":"{\"email\":\"a@b.c\",\"ticket_id\":\"99\"}"}`,
					`[{"content":"doc body one"},{"content":"doc body two"}]`),
			},
			want: `{
				"name":"agent-chat","level":"DEFAULT","input":"hi",
				"knowledge_base":{"docs":["Closing my account","Why was my account suspended"],"count":2},
				"iterations":[{"number":1}],
				"reply":"hello"
			}`,
		},
		{
			name: "untitled knowledge-base retrieval reports count only",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name: "agent", Level: "DEFAULT", Input: json.RawMessage(`"hi"`),
			}},
			obs: []api.ObservationInfo{
				obs("it1", "process", "chain", "preparation_iteration_1", "", "", "", ""),
				obs("fs1", "it1", "span", "find_similar_documents", "", "",
					`{"query":"blob"}`, `[{"content":"a"},{"content":"b"},{"content":"c"}]`),
			},
			want: `{
				"name":"agent","level":"DEFAULT","input":"hi",
				"knowledge_base":{"count":3},
				"iterations":[{"number":1}]
			}`,
		},
		{
			name: "condition whitespace normalized and tool envelope unwrapped",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name:  "agent-kyc",
				Level: "DEFAULT",
				Input: json.RawMessage(`"{\"step\":\"classify\"}"`),
			}},
			obs: []api.ObservationInfo{
				obs("it1", "process", "chain", "preparation_iteration_1", "", "", "", ""),
				obs(
					"mg1",
					"it1",
					"chain",
					"match_guidelines",
					"",
					"",
					"",
					`{"matches":[{"condition":"Applicant data looks  PROBLEMATIC\n","score":10}]}`,
				),
				obs("ex1", "it1", "tool", "execute_tool_calls", "", "", "", ""),
				obs(
					"t1",
					"ex1",
					"tool",
					"execute_thinking",
					"",
					"",
					`{"step_id":"classify"}`,
					`{"data":{"ok":true,"output":{"doc_type":"RENT_RECEIPT"}},"metadata":{},"control":{},"canned_responses":[],"canned_response_fields":{},"guidelines":[]}`,
				),
			},
			want: `{
				"name":"agent-kyc","level":"DEFAULT","input":"{\"step\":\"classify\"}",
				"iterations":[{"number":1,
					"conditions":[{"text":"Applicant data looks PROBLEMATIC","score":10}],
					"tools":[{"name":"execute_thinking","args":{"step_id":"classify"},"result":{"ok":true,"output":{"doc_type":"RENT_RECEIPT"}}}]}]
			}`,
		},
		{
			name: "tool error is captured at tool and trace level",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name: "agent", Level: "ERROR",
				Input:  json.RawMessage(`"hi"`),
				Output: json.RawMessage(`"[\"sorry\"]"`),
			}},
			obs: []api.ObservationInfo{
				obs("it1", "root", "chain", "preparation_iteration_1", "", "", "", ""),
				obs("ex1", "it1", "tool", "execute_tool_calls", "", "", "", ""),
				obs("t1", "ex1", "tool", "create_booking", "ERROR", "upstream 500",
					`"{}"`, `{"ok":false}`),
			},
			want: `{
				"name":"agent","level":"ERROR","input":"hi","reply":"sorry",
				"iterations":[{"number":1,
					"tools":[{"name":"create_booking","args":{},"result":{"ok":false},"errored":true,"error":"upstream 500"}]}],
				"errors":["create_booking: upstream 500"]
			}`,
		},
		{
			name: "display names: chat iteration with context matches and next step",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name:   "agent-chat: 406867 (turn: 4)",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"kyc status?"`),
				Output: json.RawMessage(`"[\"checking\"]"`),
			}},
			obs: []api.ObservationInfo{
				obs(
					"kb",
					"",
					"RETRIEVER",
					"Search: Documents",
					"",
					"",
					"",
					`{"has_results":true,"article_count":2,"articles":[{"name":"Documents refusés"},{"name":"Vérification KYC"}]}`,
				),
				obs("it1", "", "CHAIN", "Iteration: 1", "", "", "", ""),
				obs(
					"ec1",
					"it1",
					"CHAIN",
					"Evaluate: Context",
					"",
					"",
					"",
					`{"match_count":2,"matches":[
					{"type":"routine","routine_id":"kyc-status-update-chat","condition":"customer asks about KYC","score":10},
					{"type":"routine_node","routine_id":"kyc-status-update-chat","step_id":"check_status","condition":"player_info available","score":10},
					{"type":"policy","id":"handoff","condition":"Always applies.","score":10}
				]}`,
				),
				obs(
					"ns1",
					"ec1",
					"GENERATION",
					"Next step: KYC Status Update",
					"",
					"",
					"",
					`{"applied_condition_id":"2","next_step_rationale":"player is authenticated, check the status"}`,
				),
			},
			want: `{
				"name":"agent-chat: 406867 (turn: 4)","level":"DEFAULT","input":"kyc status?","reply":"checking",
				"knowledge_base":{"docs":["Documents refusés","Vérification KYC"],"count":2},
				"iterations":[{"number":1,
					"routines":["kyc-status-update-chat"],
					"journey":[{"routine":"kyc-status-update-chat","step":"check_status","condition":"player_info available"}],
					"conditions":[{"text":"Always applies.","score":10}],
					"decisions":["player is authenticated, check the status"]
				}]
			}`,
		},
		{
			name: "display names: routine steps and tool execution",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name:  "agent-kyc",
				Level: "DEFAULT",
				Input: json.RawMessage(`"{\"applicantId\":\"a1\"}"`),
			}},
			obs: []api.ObservationInfo{
				obs("it2", "agent", "CHAIN", "Iteration: 2", "", "", "", ""),
				obs("es1", "it2", "CHAIN", "Evaluate: Routine steps", "", "", "", `{"matches":[
					{"type":"routine_node","routine_id":"verify-kyc-l1","step_id":"l1_name_check","condition":"rejectLabels does NOT contain WRONG_ADDRESS","score":10}
				]}`),
				obs(
					"ns1",
					"es1",
					"GENERATION",
					"Next step: Verify KYC — Level 1",
					"",
					"",
					"",
					`{"applied_condition_id":"1","next_step_rationale":"proceed to name validation"}`,
				),
				obs("ex1", "it2", "CHAIN", "Execute: Tools", "", "", "", ""),
				obs(
					"t1",
					"ex1",
					"TOOL",
					"get_applicant_data",
					"",
					"",
					`{"party_id":"302110"}`,
					`{"data":{"ok":true},"metadata":{},"control":{},"canned_responses":[],"canned_response_fields":{},"guidelines":[]}`,
				),
			},
			want: `{
				"name":"agent-kyc","level":"DEFAULT","input":"{\"applicantId\":\"a1\"}",
				"iterations":[{"number":2,
					"journey":[{"routine":"verify-kyc-l1","step":"l1_name_check","condition":"rejectLabels does NOT contain WRONG_ADDRESS"}],
					"decisions":["proceed to name validation"],
					"tools":[{"name":"get_applicant_data","args":{"party_id":"302110"},"result":{"ok":true}}]
				}]
			}`,
		},
		{
			name: "warning status surfaces alongside errors",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name: "agent-kyc", Level: "WARNING",
				Input:  json.RawMessage(`"{\"type\":\"applicantOnHold\"}"`),
				Output: json.RawMessage(`"{\"matched\":[]}"`),
			}},
			obs: []api.ObservationInfo{
				obs("wh1", "", "CHAIN", "sumsub-kyc", "WARNING", "no routine matched", "", ""),
				obs("wh2", "", "CHAIN", "silent-warning", "WARNING", "", "", ""),
			},
			want: `{
				"name":"agent-kyc","level":"WARNING",
				"input":"{\"type\":\"applicantOnHold\"}","reply":"{\"matched\":[]}",
				"errors":[
					"WARNING — sumsub-kyc: no routine matched",
					"WARNING — silent-warning: warning"
				]
			}`,
		},
		{
			name: "repeated iteration numbers keep chronological order",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name: "agent-kyc", Level: "ERROR", Input: json.RawMessage(`"go"`),
			}},
			obs: []api.ObservationInfo{
				obsAt("r2i2", "", "Iteration: 2", "2026-08-03T12:00:02+02:00"),
				obsAt("r1i1", "", "Iteration: 1", "2026-08-03T10:00:00Z"),
				obsAt("r2i1", "", "Iteration: 1", "2026-08-03T09:00:01-01:00"),
				obsAt("r1i2", "", "Iteration: 2", "2026-08-03T10:00:00.1Z"),
				obs("ex1", "r1i2", "CHAIN", "Execute: Tools", "", "", "", ""),
				obs("t1", "ex1", "TOOL", "emit_output", "ERROR", "output_validation_failed",
					`{}`, `{"ok":false}`),
			},
			want: `{
				"name":"agent-kyc","level":"ERROR","input":"go",
				"iterations":[
					{"number":1},
					{"number":2,"tools":[{"name":"emit_output","args":{},"result":{"ok":false},"errored":true,"error":"output_validation_failed"}]},
					{"number":1},
					{"number":2}
				],
				"errors":["emit_output: output_validation_failed"]
			}`,
		},
		{
			name: "no observations still renders input and reply",
			trace: &api.TraceDetail{TraceInfo: api.TraceInfo{
				Name: "agent", Level: "ERROR",
				Input:  json.RawMessage(`"hi"`),
				Output: json.RawMessage(`"[\"sorry\"]"`),
			}},
			obs:  nil,
			want: `{"name":"agent","level":"ERROR","input":"hi","reply":"sorry"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertJSON(t, TraceSummary(tc.trace, tc.obs), tc.want)
		})
	}
}

func TestUnwrapToolResult(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			"envelope collapses to data",
			`{"data":{"ok":true,"value":3},"metadata":{},"control":{},"canned_responses":[],"canned_response_fields":{},"guidelines":[]}`,
			`{"ok":true,"value":3}`,
		},
		{
			"string-wrapped envelope",
			`"{\"data\":{\"x\":1},\"metadata\":{},\"control\":{}}"`,
			`{"x":1}`,
		},
		{
			"unexpected sibling passes through",
			`{"data":{"x":1},"other":true}`,
			`{"data":{"x":1},"other":true}`,
		},
		{"no data key passes through", `{"count":3}`, `{"count":3}`},
		{"non-object passes through", `[1,2]`, `[1,2]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CompactJSON(UnwrapToolResult(json.RawMessage(tc.raw)))
			if got != tc.want {
				t.Fatalf("UnwrapToolResult(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestTraceSummary_CyclicGraph(t *testing.T) {
	trace := &api.TraceDetail{TraceInfo: api.TraceInfo{
		Name: "agent", Level: "DEFAULT", Input: json.RawMessage(`"hi"`),
	}}
	// Malformed tree: a subtree with a parent cycle (a -> b -> a) and a
	// self-reference (c -> c). The summary must terminate without panic.
	observations := []api.ObservationInfo{
		obs("it1", "process", "chain", "preparation_iteration_1", "", "", "", ""),
		obs("a", "it1", "span", "node_a", "", "", "", ""),
		obs("b", "a", "span", "node_b", "", "", "", ""),
		obs("a", "b", "span", "node_a_cycle", "", "", "", ""), // b -> a back-edge
		obs("c", "it1", "span", "node_c", "", "", "", ""),
		obs("c", "c", "span", "node_c_self", "", "", "", ""), // c -> c self-loop
	}

	done := make(chan *TraceSummaryModel, 1)
	go func() { done <- TraceSummary(trace, observations) }()
	select {
	case m := <-done:
		if m == nil || len(m.Iterations) != 1 {
			t.Fatalf("cyclic summary = %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("TraceSummary did not terminate on a cyclic observation graph")
	}
}
