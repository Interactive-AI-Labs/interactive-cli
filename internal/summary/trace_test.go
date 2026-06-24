package summary

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func obs(id, parent, typ, name, level, status string, in, out string) clients.ObservationInfo {
	o := clients.ObservationInfo{
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

// assertJSON marshals got and compares it to wantJSON structurally (ignoring key
// order and whitespace), so each case can state its expected summary model as the
// JSON `--summary --json` would emit.
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
		trace *clients.TraceDetail
		obs   []clients.ObservationInfo
		want  string
	}{
		{
			name: "two iterations with conditions and a tool",
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
				Name:   "driveaway-agent",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"I want to rent a car for next weekend"`),
				Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
			}},
			obs: []clients.ObservationInfo{
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
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
				Name:   "agent-chat",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"first bet refund?"`),
				Output: json.RawMessage(`"[\"transfer\"]"`),
			}},
			obs: []clients.ObservationInfo{
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
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
				Name:   "agent-chat",
				Level:  "DEFAULT",
				Input:  json.RawMessage(`"hi"`),
				Output: json.RawMessage(`"[\"hello\"]"`),
			}},
			obs: []clients.ObservationInfo{
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
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
				Name: "agent", Level: "DEFAULT", Input: json.RawMessage(`"hi"`),
			}},
			obs: []clients.ObservationInfo{
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
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
				Name:  "agent-kyc",
				Level: "DEFAULT",
				Input: json.RawMessage(`"{\"step\":\"classify\"}"`),
			}},
			obs: []clients.ObservationInfo{
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
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
				Name: "agent", Level: "ERROR",
				Input:  json.RawMessage(`"hi"`),
				Output: json.RawMessage(`"[\"sorry\"]"`),
			}},
			obs: []clients.ObservationInfo{
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
			name: "no observations still renders input and reply",
			trace: &clients.TraceDetail{TraceInfo: clients.TraceInfo{
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
	trace := &clients.TraceDetail{TraceInfo: clients.TraceInfo{
		Name: "agent", Level: "DEFAULT", Input: json.RawMessage(`"hi"`),
	}}
	// Malformed tree: a subtree with a parent cycle (a -> b -> a) and a
	// self-reference (c -> c). The summary must terminate without panic.
	observations := []clients.ObservationInfo{
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
