package summary

import (
	"encoding/json"
	"testing"

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
			Name: "driveaway-agent", Level: "DEFAULT",
		},
		Input:  json.RawMessage(`"I want to rent a car for next weekend"`),
		Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
	}
	observations := []clients.ObservationInfo{
		obs("it1", "root", "chain", "preparation_iteration_1", "", "", "", ""),
		obs("mg1", "it1", "chain", "match_guidelines", "", "",
			"", `{"matches":[{"condition":"Customer asks to rent a vehicle","score":9},{"condition":"No booking in progress","score":7}]}`),
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
	if got := m.Iterations[0].Conditions; len(got) != 2 || got[0].Text != "Customer asks to rent a vehicle" || got[0].Score != 9 {
		t.Fatalf("iter1 conditions = %+v", got)
	}
	if got := m.Iterations[0].Tools; len(got) != 1 || got[0].Name != "check_availability" || got[0].Args != `dates="next weekend"` {
		t.Fatalf("iter1 tools = %+v", got)
	}
	if len(m.Iterations[1].Tools) != 0 {
		t.Fatalf("iter2 should have no tools, got %+v", m.Iterations[1].Tools)
	}
}

func TestTraceSummary_ToolErrorAndNoObservations(t *testing.T) {
	trace := &clients.TraceDetail{
		TraceInfo: clients.TraceInfo{Name: "agent", Level: "ERROR"},
		Input:     json.RawMessage(`"hi"`),
		Output:    json.RawMessage(`"[\"sorry\"]"`),
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
