package summary

import (
	"encoding/json"
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

func TestTraceSummary_TwoIterations(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			Name:   "driveaway-agent",
			Level:  "DEFAULT",
			Input:  json.RawMessage(`"I want to rent a car for next weekend"`),
			Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
		},
	}
	observations := []clients.ObservationInfo{
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
	}

	m := TraceSummary(trace, observations)

	if m.Input != "I want to rent a car for next weekend" {
		t.Fatalf("Input = %q", m.Input)
	}
	if m.Reply != "Great! We have 3 cars available..." {
		t.Fatalf("Reply = %q", m.Reply)
	}
	if len(m.Iterations) != 2 {
		t.Fatalf("want 2 iterations, got %d", len(m.Iterations))
	}
	if got := m.Iterations[0].Conditions; len(got) != 2 ||
		got[0].Text != "Customer asks to rent a vehicle" ||
		got[0].Score != 9 {
		t.Fatalf("iter1 conditions = %+v", got)
	}
	if got := m.Iterations[0].Tools; len(got) != 1 || got[0].Name != "check_availability" ||
		got[0].Args != `dates="next weekend"` {
		t.Fatalf("iter1 tools = %+v", got)
	}
	if len(m.Iterations[1].Tools) != 0 {
		t.Fatalf("iter2 should have no tools, got %+v", m.Iterations[1].Tools)
	}
}

func TestTraceSummary_KnowledgeBase(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			Name:   "agent-chat",
			Level:  "DEFAULT",
			Input:  json.RawMessage(`"hi"`),
			Output: json.RawMessage(`"[\"hello\"]"`),
		},
	}
	observations := []clients.ObservationInfo{
		// Titled retriever lives at the root, emitted once for the turn.
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
		// Per-iteration untitled vector searches: their context-blob query must NOT leak.
		obs("fs1", "it1", "span", "find_similar_documents", "", "",
			`{"query":"{\"email\":\"a@b.c\",\"ticket_id\":\"99\"}"}`,
			`[{"content":"doc body one"},{"content":"doc body two"}]`),
	}

	m := TraceSummary(trace, observations)
	if m.KB == nil {
		t.Fatalf("expected KB summary")
	}
	if len(m.KB.Docs) != 2 || m.KB.Docs[0] != "Closing my account" ||
		m.KB.Docs[1] != "Why was my account suspended" {
		t.Fatalf("KB docs = %+v", m.KB.Docs)
	}
	if m.KB.Count != 2 {
		t.Fatalf("KB count = %d, want 2", m.KB.Count)
	}
}

func TestTraceSummary_KnowledgeBaseUntitled(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			Name:  "agent",
			Level: "DEFAULT",
			Input: json.RawMessage(`"hi"`),
		},
	}
	observations := []clients.ObservationInfo{
		obs("it1", "process", "chain", "preparation_iteration_1", "", "", "", ""),
		obs("fs1", "it1", "span", "find_similar_documents", "", "",
			`{"query":"blob"}`, `[{"content":"a"},{"content":"b"},{"content":"c"}]`),
	}
	m := TraceSummary(trace, observations)
	if m.KB == nil || len(m.KB.Docs) != 0 || m.KB.Count != 3 {
		t.Fatalf("untitled KB = %+v", m.KB)
	}
}

func TestTraceSummary_ConditionNormalizedAndToolEnvelope(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			Name:  "agent-kyc",
			Level: "DEFAULT",
			Input: json.RawMessage(`"{\"step\":\"classify\"}"`),
		},
	}
	observations := []clients.ObservationInfo{
		obs("it1", "process", "chain", "preparation_iteration_1", "", "", "", ""),
		obs("mg1", "it1", "chain", "match_guidelines", "", "",
			"", `{"matches":[{"condition":"Applicant data looks  PROBLEMATIC\n","score":10}]}`),
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
	}
	m := TraceSummary(trace, observations)
	if len(m.Iterations) != 1 {
		t.Fatalf("iterations = %+v", m.Iterations)
	}
	cond := m.Iterations[0].Conditions
	if len(cond) != 1 || cond[0].Text != "Applicant data looks PROBLEMATIC" {
		t.Fatalf("condition not whitespace-normalized: %+v", cond)
	}
	tool := m.Iterations[0].Tools
	if len(tool) != 1 || tool[0].Result != `{"ok":true,"output":{"doc_type":"RENT_RECEIPT"}}` {
		t.Fatalf("tool result envelope not unwrapped: %+v", tool)
	}
}

func TestTraceSummary_CyclicGraph(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			Name:  "agent",
			Level: "DEFAULT",
			Input: json.RawMessage(`"hi"`),
		},
	}
	// Malformed tree: an iteration whose subtree contains a parent cycle
	// (a -> b -> a) plus a self-reference (c -> c). Must not loop forever.
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

func TestTraceSummary_ToolErrorAndNoObservations(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{
			Name:   "agent",
			Level:  "ERROR",
			Input:  json.RawMessage(`"hi"`),
			Output: json.RawMessage(`"[\"sorry\"]"`),
		},
	}
	// tool error
	observations := []clients.ObservationInfo{
		obs("it1", "root", "chain", "preparation_iteration_1", "", "", "", ""),
		obs("ex1", "it1", "tool", "execute_tool_calls", "", "", "", ""),
		obs("t1", "ex1", "tool", "create_booking", "ERROR", "upstream 500",
			`"{}"`, `{"ok":false}`),
	}
	m := TraceSummary(trace, observations)
	if len(m.Iterations) != 1 || len(m.Iterations[0].Tools) != 1 {
		t.Fatalf("iterations = %+v", m.Iterations)
	}
	tc := m.Iterations[0].Tools[0]
	if !tc.Errored || tc.ErrMsg != "upstream 500" {
		t.Fatalf("tool err = %+v", tc)
	}
	if len(m.Errors) == 0 {
		t.Fatalf("expected trace-level Errors collected")
	}

	// no observations: still renders input + reply, zero iterations
	m2 := TraceSummary(trace, nil)
	if m2.Input != "hi" || m2.Reply != "sorry" || len(m2.Iterations) != 0 {
		t.Fatalf("empty-obs summary = %+v", m2)
	}
}
