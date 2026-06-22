# LLM-Readable Trace & Session Summaries Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an additive `--summary` flag to `iai traces get` and `iai sessions get` that renders compact, token-lean, LLM-readable narratives reconstructed from existing trace/observation data.

**Architecture:** A new pure-transform package `internal/summary/` turns the typed client structs (`*clients.TraceDetail` + `[]clients.ObservationInfo` for a turn; `[]clients.TraceInfo` for a session) into clean summary models. Thin renderers in `internal/output/` print those models. The two `cmd/` files gain a `--summary` flag that branches before the existing JSON/YAML/detail output. One small additive change to `clients.TraceInfo` (two `omitempty` IO fields) lets the session view fetch per-turn messages in a single `ListTraces` call.

**Tech Stack:** Go, cobra, `encoding/json`, `text/tabwriter`. Tests use the standard `testing` package (table-driven), matching existing `internal/**/*_test.go` conventions.

## Global Constraints

- **Strictly additive.** Do not remove or alter any existing command, flag, column, or output path. `--json`/`--yaml`/`--columns`/`--fields` behavior is unchanged when `--summary` is absent.
- **Scope: CLI only.** No platform/API changes. No new client method signatures except adding two `omitempty` fields to `TraceInfo`.
- **`--summary` is mutually exclusive with `--json` and `--yaml`.**
- **Token discipline:** truncate long values to 500 chars (trace view) / 160 chars (session messages) with a trailing `… (truncated)`; rune-safe. Guideline rationale omitted by default.
- **Robustness:** unrecognized spans are skipped, never panic; IO that is a JSON-encoded string wrapping JSON must be unwrapped (Langfuse stores some IO this way — mirrors the existing `prettyJSONUnwrapString` quirk in `internal/output/traces.go`).
- **Go module path:** `github.com/Interactive-AI-Labs/interactive-cli`.
- Run `gofmt`/existing lint before each commit; the repo uses tabs (Go default).

---

## File Structure

- Create `internal/summary/models.go` — model structs (Task 1)
- Create `internal/summary/value.go` — JSON value/truncation helpers (Task 1)
- Create `internal/summary/value_test.go` (Task 1)
- Create `internal/summary/trace.go` — `TraceSummary(...)` (Task 2)
- Create `internal/summary/trace_test.go` (Task 2)
- Create `internal/summary/session.go` — `SessionSummary(...)` (Task 3)
- Create `internal/summary/session_test.go` (Task 3)
- Create `internal/output/summary_traces.go` — `PrintTraceSummary(...)` (Task 4)
- Create `internal/output/summary_traces_test.go` (Task 4)
- Create `internal/output/summary_sessions.go` — `PrintSessionSummary(...)` (Task 5)
- Create `internal/output/summary_sessions_test.go` (Task 5)
- Modify `cmd/traces.go` — add `--summary` to `traces get` (Task 6)
- Modify `internal/clients/api_client.go:357-376` — add IO fields to `TraceInfo` (Task 7)
- Modify `cmd/sessions.go` — add `--summary` to `sessions get` (Task 7)
- Regenerate `docs/` via `make docs` (Task 8)

---

### Task 1: Summary models + JSON value helpers

**Files:**
- Create: `internal/summary/models.go`
- Create: `internal/summary/value.go`
- Test: `internal/summary/value_test.go`

**Interfaces:**
- Consumes: nothing (leaf package, stdlib only).
- Produces (used by Tasks 2–5):
  - Types `Condition{Text string; Score int}`, `ToolCall{Name, Args, Result string; Errored bool; ErrMsg string}`, `Iteration{Number int; Conditions []Condition; Tools []ToolCall; KBQueries []string}`, `TraceSummaryModel{Name, Timestamp string; LatencyMs, Cost *float64; Level, Input string; Iterations []Iteration; Reply string; Errors []string}`, `Turn{Number int; Customer, Agent string; Tools, Journeys []string}`, `SessionSummaryModel{ID, Agent string; TurnCount int; Duration string; Cost *float64; Turns []Turn}`.
  - Helpers `AsString(raw json.RawMessage) string`, `UnwrapJSON(raw json.RawMessage) json.RawMessage`, `Truncate(s string, max int) string`, `CollapseWS(s string) string`, `CompactJSON(raw json.RawMessage) string`, `CompactArgs(raw json.RawMessage) string`.
  - Constants `MaxValueLen = 500`, `MaxSessionMsgLen = 160`.

- [ ] **Step 1: Write the failing test**

Create `internal/summary/value_test.go`:

```go
package summary

import (
	"encoding/json"
	"testing"
)

func TestAsString(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"plain json string", `"hello world"`, "hello world"},
		{"string-wrapped array", `"[\"a\",\"b\"]"`, "a\nb"},
		{"native array", `["x","y"]`, "x\ny"},
		{"object falls back to compact json", `{"k":1}`, `{"k":1}`},
		{"empty", ``, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := AsString(json.RawMessage(tc.raw))
			if got != tc.want {
				t.Fatalf("AsString(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestUnwrapJSON(t *testing.T) {
	// A JSON string whose content is a JSON object must unwrap to the object.
	got := UnwrapJSON(json.RawMessage(`"{\"a\":1}"`))
	var m map[string]int
	if err := json.Unmarshal(got, &m); err != nil || m["a"] != 1 {
		t.Fatalf("UnwrapJSON did not unwrap string-wrapped object: %s (err %v)", got, err)
	}
	// A native object passes through.
	got = UnwrapJSON(json.RawMessage(`{"b":2}`))
	if err := json.Unmarshal(got, &m); err != nil || m["b"] != 2 {
		t.Fatalf("UnwrapJSON mangled native object: %s", got)
	}
}

func TestTruncateRuneSafe(t *testing.T) {
	if got := Truncate("  abcdef  ", 3); got != "abc… (truncated)" {
		t.Fatalf("Truncate = %q", got)
	}
	if got := Truncate("héllo", 10); got != "héllo" {
		t.Fatalf("Truncate short = %q", got)
	}
}

func TestCompactArgs(t *testing.T) {
	got := CompactArgs(json.RawMessage(`"{\"dates\":\"next weekend\",\"qty\":2}"`))
	// keys sorted, k=v form
	if got != `dates="next weekend", qty=2` {
		t.Fatalf("CompactArgs = %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/summary/ -run 'TestAsString|TestUnwrapJSON|TestTruncate|TestCompactArgs' -v`
Expected: FAIL — `internal/summary` package does not compile (undefined: AsString, …).

- [ ] **Step 3: Write `internal/summary/models.go`**

```go
// Package summary turns raw trace/observation data into compact, LLM-readable
// summary models. It performs no I/O and no rendering.
package summary

// Condition is a guideline whose condition matched ("marked true") in a turn.
type Condition struct {
	Text  string
	Score int
}

// ToolCall is a single tool invocation within an iteration.
type ToolCall struct {
	Name    string
	Args    string
	Result  string
	Errored bool
	ErrMsg  string
}

// Iteration is one engine preparation iteration within a turn.
type Iteration struct {
	Number     int
	Conditions []Condition
	Tools      []ToolCall
	KBQueries  []string
}

// TraceSummaryModel is the compact view of a single turn (one trace).
type TraceSummaryModel struct {
	Name       string
	Timestamp  string
	LatencyMs  *float64
	Cost       *float64
	Level      string
	Input      string
	Iterations []Iteration
	Reply      string
	Errors     []string
}

// Turn is one back-and-forth within a session.
type Turn struct {
	Number   int
	Customer string
	Agent    string
	Tools    []string
	Journeys []string
}

// SessionSummaryModel is the overall view of a conversation (one session).
type SessionSummaryModel struct {
	ID        string
	Agent     string
	TurnCount int
	Duration  string
	Cost      *float64
	Turns     []Turn
}
```

- [ ] **Step 4: Write `internal/summary/value.go`**

```go
package summary

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	// MaxValueLen caps tool args/results and trace input/reply.
	MaxValueLen = 500
	// MaxSessionMsgLen caps per-turn messages in the session (overview) view.
	MaxSessionMsgLen = 160
)

// UnwrapJSON removes one layer of JSON-string wrapping if the raw value is a
// JSON string whose contents are themselves valid JSON. Langfuse stores some
// IO this way (e.g. "{\"a\":1}"). A native JSON value passes through unchanged.
func UnwrapJSON(raw json.RawMessage) json.RawMessage {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var inner json.RawMessage
		if json.Unmarshal([]byte(s), &inner) == nil {
			return inner
		}
		b, _ := json.Marshal(s) // plain string: re-encode as valid JSON
		return b
	}
	return raw
}

// AsString renders a raw IO value as human text:
//   - JSON string            -> the string (recursing if it wraps an array)
//   - JSON array of strings  -> joined by newlines
//   - anything else          -> compact JSON
func AsString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if joined, ok := tryJoinStringArray([]byte(s)); ok {
			return joined
		}
		return s
	}
	if joined, ok := tryJoinStringArray(raw); ok {
		return joined
	}
	return CompactJSON(raw)
}

func tryJoinStringArray(b []byte) (string, bool) {
	var arr []string
	if err := json.Unmarshal(b, &arr); err == nil {
		return strings.Join(arr, "\n"), true
	}
	return "", false
}

// CompactJSON re-encodes raw JSON without indentation; on failure returns the
// raw bytes as a string.
func CompactJSON(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}
	return string(b)
}

// CompactArgs renders tool arguments as `k=v, k=v` (keys sorted) for a flat
// object, falling back to compact JSON for anything else.
func CompactArgs(raw json.RawMessage) string {
	obj := UnwrapJSON(raw)
	var m map[string]any
	if err := json.Unmarshal(obj, &m); err != nil {
		return CompactJSON(obj)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, scalar(m[k])))
	}
	return strings.Join(parts, ", ")
}

func scalar(v any) string {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("%q", t)
	case float64:
		// JSON numbers decode to float64; print integers without trailing zeros.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case nil:
		return "null"
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// Truncate trims whitespace and caps s to max runes, appending a marker.
func Truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "… (truncated)"
}

// CollapseWS collapses all runs of whitespace (incl. newlines) to single spaces.
func CollapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/summary/ -v`
Expected: PASS (all value tests).

- [ ] **Step 6: Commit**

```bash
git add internal/summary/models.go internal/summary/value.go internal/summary/value_test.go
git commit -m "feat(summary): add summary models and JSON value helpers"
```

---

### Task 2: Trace extraction (`TraceSummary`)

**Files:**
- Create: `internal/summary/trace.go`
- Test: `internal/summary/trace_test.go`

**Interfaces:**
- Consumes: `internal/summary` helpers/models (Task 1); `clients.TraceDetail`, `clients.ObservationInfo` from `github.com/Interactive-AI-Labs/interactive-cli/internal/clients`.
- Produces (used by Task 6): `func TraceSummary(trace *clients.TraceDetail, obs []clients.ObservationInfo) *TraceSummaryModel`.

Note on data shapes (verified against interactive-agent instrumentation):
- Iteration observations are named `preparation_iteration_N`.
- `match_guidelines` observation `Output` = `{"matches":[{"condition":..., "score":int, ...}]}` (may be a JSON-string-wrapped object). A `match_guidelines` span may appear more than once per iteration (initial + reevaluation passes) and is nested several levels below the iteration span, so search the whole iteration subtree and aggregate.
- `execute_tool_calls` observation's **direct children** are the per-tool spans (observation `Name` = tool name, `Input` = args, `Output` = result; errored → `Level == "ERROR"`, message in `StatusMessage`).
- KB spans are named `retriever:knowledge_base` or `find_similar_documents`.
- Customer message = `trace.Input`; agent reply = `trace.Output`.

- [ ] **Step 1: Write the failing test**

Create `internal/summary/trace_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/summary/ -run TestTraceSummary -v`
Expected: FAIL — `undefined: TraceSummary`.

- [ ] **Step 3: Write `internal/summary/trace.go`**

```go
package summary

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var iterationNameRe = regexp.MustCompile(`^preparation_iteration_(\d+)$`)

type matchOutput struct {
	Matches []struct {
		Condition string `json:"condition"`
		Score     int    `json:"score"`
	} `json:"matches"`
}

// TraceSummary reconstructs the observation tree for one turn and extracts the
// per-iteration conditions, tools, KB queries, the agent reply, and any errors.
func TraceSummary(trace *clients.TraceDetail, obs []clients.ObservationInfo) *TraceSummaryModel {
	m := &TraceSummaryModel{
		Name:      trace.Name,
		Timestamp: trace.Timestamp,
		LatencyMs: trace.LatencyMs,
		Cost:      trace.TotalCost,
		Level:     trace.Level,
		Input:     Truncate(AsString(trace.Input), MaxValueLen),
		Reply:     Truncate(AsString(trace.Output), MaxValueLen),
	}

	byID := make(map[string]clients.ObservationInfo, len(obs))
	children := make(map[string][]clients.ObservationInfo, len(obs))
	for _, o := range obs {
		byID[o.ID] = o
		children[o.ParentObservationID] = append(children[o.ParentObservationID], o)
		if strings.EqualFold(o.Level, "ERROR") {
			msg := o.StatusMessage
			if msg == "" {
				msg = "error"
			}
			m.Errors = append(m.Errors, o.Name+": "+msg)
		}
	}

	// Collect iteration observations, sorted by their numeric suffix.
	type iterNode struct {
		num int
		id  string
	}
	var iters []iterNode
	for _, o := range obs {
		if sm := iterationNameRe.FindStringSubmatch(o.Name); sm != nil {
			n, _ := strconv.Atoi(sm[1])
			iters = append(iters, iterNode{num: n, id: o.ID})
		}
	}
	sort.Slice(iters, func(i, j int) bool { return iters[i].num < iters[j].num })

	for _, it := range iters {
		iteration := Iteration{Number: it.num}
		sub := descendants(children, it.id)

		// Conditions: aggregate across all match_guidelines spans in subtree,
		// dedup by condition text (keep highest score).
		condScore := map[string]int{}
		var condOrder []string
		for _, d := range sub {
			if d.Name != "match_guidelines" || len(d.Output) == 0 {
				continue
			}
			var mo matchOutput
			if json.Unmarshal(UnwrapJSON(d.Output), &mo) != nil {
				continue
			}
			for _, mm := range mo.Matches {
				if mm.Condition == "" {
					continue
				}
				if prev, ok := condScore[mm.Condition]; !ok || mm.Score > prev {
					if !ok {
						condOrder = append(condOrder, mm.Condition)
					}
					condScore[mm.Condition] = mm.Score
				}
			}
		}
		for _, c := range condOrder {
			iteration.Conditions = append(iteration.Conditions, Condition{Text: c, Score: condScore[c]})
		}

		// Tools: direct children of any execute_tool_calls span in subtree.
		for _, d := range sub {
			if d.Name != "execute_tool_calls" {
				continue
			}
			for _, tool := range children[d.ID] {
				tc := ToolCall{
					Name:   tool.Name,
					Args:   Truncate(CompactArgs(tool.Input), MaxValueLen),
					Result: Truncate(CompactJSON(UnwrapJSON(tool.Output)), MaxValueLen),
				}
				if strings.EqualFold(tool.Level, "ERROR") {
					tc.Errored = true
					tc.ErrMsg = tool.StatusMessage
				}
				iteration.Tools = append(iteration.Tools, tc)
			}
		}

		// KB queries.
		for _, d := range sub {
			if d.Name == "retriever:knowledge_base" || d.Name == "find_similar_documents" {
				if q := strings.TrimSpace(AsString(d.Input)); q != "" {
					iteration.KBQueries = append(iteration.KBQueries, Truncate(CollapseWS(q), MaxValueLen))
				}
			}
		}

		m.Iterations = append(m.Iterations, iteration)
	}

	return m
}

// descendants returns all transitive children of id (excluding id itself).
func descendants(children map[string][]clients.ObservationInfo, id string) []clients.ObservationInfo {
	var out []clients.ObservationInfo
	stack := []string{id}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, c := range children[cur] {
			out = append(out, c)
			stack = append(stack, c.ID)
		}
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/summary/ -run TestTraceSummary -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/summary/trace.go internal/summary/trace_test.go
git commit -m "feat(summary): reconstruct per-iteration trace summary from observations"
```

---

### Task 3: Session extraction (`SessionSummary`)

**Files:**
- Create: `internal/summary/session.go`
- Test: `internal/summary/session_test.go`

**Interfaces:**
- Consumes: `internal/summary` helpers/models (Task 1); `clients.TraceInfo` (with `Input`/`Output` IO fields added in Task 7 — until then the test sets them via the fields added in Task 7; **Task 7 must land the struct fields, so do Task 7's Step "add IO fields to TraceInfo" before running this task's tests** — see note below).
- Produces (used by Task 7): `func SessionSummary(sessionID string, traces []clients.TraceInfo) *SessionSummaryModel`.

> **Ordering note:** `SessionSummary` reads `clients.TraceInfo.Input` / `.Output`. Those fields are added in Task 7, Step 1. To keep this task self-contained and compilable, **move Task 7's "add IO fields to TraceInfo" step to the front of this task** (do it as Step 0 here) and drop it from Task 7. The plan keeps the edit described in one place (Task 7) for clarity; perform it here first if implementing in order.

- [ ] **Step 0: Ensure `clients.TraceInfo` has IO fields** (see Task 7, Step 1 for the exact edit). After editing, `go build ./...` must succeed.

- [ ] **Step 1: Write the failing test**

Create `internal/summary/session_test.go`:

```go
package summary

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestSessionSummary(t *testing.T) {
	traces := []clients.TraceInfo{
		{
			ID: "t1", Name: "turn", Timestamp: "2026-06-22T14:30:00Z",
			Tags:   []string{"agent:driveaway-agent", "tool:check_availability"},
			Input:  json.RawMessage(`"I want to rent a car next weekend"`),
			Output: json.RawMessage(`"[\"Great! We have 3 cars available...\"]"`),
		},
		{
			ID: "t2", Name: "turn", Timestamp: "2026-06-22T14:32:00Z",
			Tags:   []string{"agent:driveaway-agent", "tool:create_booking", "routine:rental"},
			Input:  json.RawMessage(`"Downtown"`),
			Output: json.RawMessage(`"[\"Booked! Confirmation #1234\"]"`),
		},
	}

	m := SessionSummary("s_abc", traces)

	if m.ID != "s_abc" || m.Agent != "driveaway-agent" || m.TurnCount != 2 {
		t.Fatalf("header = %+v", m)
	}
	if m.Duration != "2m0s" {
		t.Fatalf("Duration = %q", m.Duration)
	}
	if len(m.Turns) != 2 {
		t.Fatalf("turns = %d", len(m.Turns))
	}
	if m.Turns[0].Customer != "I want to rent a car next weekend" ||
		m.Turns[0].Agent != "Great! We have 3 cars available..." {
		t.Fatalf("turn1 = %+v", m.Turns[0])
	}
	if len(m.Turns[0].Tools) != 1 || m.Turns[0].Tools[0] != "check_availability" {
		t.Fatalf("turn1 tools = %+v", m.Turns[0].Tools)
	}
	if len(m.Turns[1].Journeys) != 1 || m.Turns[1].Journeys[0] != "rental" {
		t.Fatalf("turn2 journeys = %+v", m.Turns[1].Journeys)
	}
}

func TestSessionSummary_Empty(t *testing.T) {
	m := SessionSummary("s_x", nil)
	if m.ID != "s_x" || m.TurnCount != 0 || len(m.Turns) != 0 {
		t.Fatalf("empty session = %+v", m)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/summary/ -run TestSessionSummary -v`
Expected: FAIL — `undefined: SessionSummary`.

- [ ] **Step 3: Write `internal/summary/session.go`**

```go
package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// SessionSummary builds the conversation overview from a session's traces
// (one trace per turn). Traces are ordered by timestamp ascending.
func SessionSummary(sessionID string, traces []clients.TraceInfo) *SessionSummaryModel {
	m := &SessionSummaryModel{ID: sessionID, TurnCount: len(traces)}

	sorted := make([]clients.TraceInfo, len(traces))
	copy(sorted, traces)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Timestamp < sorted[j].Timestamp
	})

	var costSum float64
	var haveCost bool
	for i, tr := range sorted {
		if tr.TotalCost != nil {
			costSum += *tr.TotalCost
			haveCost = true
		}
		turn := Turn{
			Number:   i + 1,
			Customer: Truncate(CollapseWS(AsString(tr.Input)), MaxSessionMsgLen),
			Agent:    Truncate(CollapseWS(AsString(tr.Output)), MaxSessionMsgLen),
		}
		for _, tag := range tr.Tags {
			switch {
			case strings.HasPrefix(tag, "tool:"):
				turn.Tools = append(turn.Tools, strings.TrimPrefix(tag, "tool:"))
			case strings.HasPrefix(tag, "routine:"):
				turn.Journeys = append(turn.Journeys, strings.TrimPrefix(tag, "routine:"))
			case strings.HasPrefix(tag, "agent:") && m.Agent == "":
				m.Agent = strings.TrimPrefix(tag, "agent:")
			}
		}
		m.Turns = append(m.Turns, turn)
	}

	if haveCost {
		m.Cost = &costSum
	}
	m.Duration = sessionDuration(sorted)
	return m
}

func sessionDuration(sorted []clients.TraceInfo) string {
	if len(sorted) < 2 {
		return ""
	}
	first, err1 := parseTS(sorted[0].Timestamp)
	last, err2 := parseTS(sorted[len(sorted)-1].Timestamp)
	if err1 != nil || err2 != nil || last.Before(first) {
		return ""
	}
	d := last.Sub(first).Round(time.Second)
	return d.String()
}

func parseTS(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable timestamp %q", s)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/summary/ -v`
Expected: PASS (all summary tests). Note: `time.Duration("2m0s")` — `(2*time.Minute).String()` is `"2m0s"`, matching the test.

- [ ] **Step 5: Commit**

```bash
git add internal/summary/session.go internal/summary/session_test.go internal/clients/api_client.go
git commit -m "feat(summary): build session conversation overview from traces"
```

---

### Task 4: Trace summary renderer (`PrintTraceSummary`)

**Files:**
- Create: `internal/output/summary_traces.go`
- Test: `internal/output/summary_traces_test.go`

**Interfaces:**
- Consumes: `summary.TraceSummaryModel` (Task 2); existing helpers `LocalTime`, `formatLatencyMs`, `formatCost` (in `internal/output`).
- Produces (used by Task 6): `func PrintTraceSummary(out io.Writer, m *summary.TraceSummaryModel) error`.

- [ ] **Step 1: Write the failing test**

Create `internal/output/summary_traces_test.go`:

```go
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
				Number:     1,
				Conditions: []summary.Condition{{Text: "Customer asks to rent a vehicle", Score: 9}},
				Tools:      []summary.ToolCall{{Name: "check_availability", Args: `dates="next weekend"`, Result: `{"count":3}`}},
			},
			{
				Number:     2,
				Conditions: []summary.Condition{{Text: "Pickup location not yet provided", Score: 8}},
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

func TestPrintTraceSummary_ErrorAndToolError(t *testing.T) {
	m := &summary.TraceSummaryModel{
		Name:  "agent",
		Level: "ERROR",
		Input: "hi",
		Iterations: []summary.Iteration{
			{Number: 1, Tools: []summary.ToolCall{{Name: "create_booking", Args: "", Errored: true, ErrMsg: "upstream 500"}}},
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/output/ -run TestPrintTraceSummary -v`
Expected: FAIL — `undefined: PrintTraceSummary`.

- [ ] **Step 3: Write `internal/output/summary_traces.go`**

```go
package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

// PrintTraceSummary renders one turn as a compact, LLM-readable narrative.
func PrintTraceSummary(out io.Writer, m *summary.TraceSummaryModel) error {
	var b strings.Builder

	header := fmt.Sprintf("Turn — %s", m.Name)
	if ts := LocalTime(m.Timestamp); ts != "" && m.Timestamp != "" {
		header += " · " + ts
	}
	if m.LatencyMs != nil {
		header += " · " + formatLatencyMs(m.LatencyMs)
	}
	if m.Cost != nil {
		header += " · " + formatCost(m.Cost)
	}
	if strings.EqualFold(m.Level, "ERROR") {
		header += " · ERROR"
	}
	header += fmt.Sprintf(" · %d iterations", len(m.Iterations))
	b.WriteString(header + "\n\n")

	if m.Input != "" {
		b.WriteString("Customer: " + m.Input + "\n\n")
	}

	for _, it := range m.Iterations {
		b.WriteString(fmt.Sprintf("Iteration %d\n", it.Number))
		if len(it.Conditions) > 0 {
			b.WriteString("  Conditions met:\n")
			for _, c := range it.Conditions {
				b.WriteString(fmt.Sprintf("    ✓ %s (%d)\n", c.Text, c.Score))
			}
		}
		for _, q := range it.KBQueries {
			b.WriteString("  Knowledge base: " + q + "\n")
		}
		if len(it.Tools) > 0 {
			b.WriteString("  Tools called:\n")
			for _, tc := range it.Tools {
				line := fmt.Sprintf("    → %s(%s)", tc.Name, tc.Args)
				if tc.Errored {
					msg := tc.ErrMsg
					if msg == "" {
						msg = "error"
					}
					line += " → ERROR: " + msg
				} else if tc.Result != "" {
					line += " → " + tc.Result
				}
				b.WriteString(line + "\n")
			}
		} else {
			b.WriteString("  (no tools called)\n")
		}
	}

	if m.Reply != "" {
		b.WriteString("\nAgent: " + m.Reply + "\n")
	}

	if len(m.Errors) > 0 {
		b.WriteString("\nErrors:\n")
		for _, e := range m.Errors {
			b.WriteString("  - " + e + "\n")
		}
	}

	_, err := io.WriteString(out, b.String())
	return err
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/output/ -run TestPrintTraceSummary -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/output/summary_traces.go internal/output/summary_traces_test.go
git commit -m "feat(output): render compact trace summary"
```

---

### Task 5: Session summary renderer (`PrintSessionSummary`)

**Files:**
- Create: `internal/output/summary_sessions.go`
- Test: `internal/output/summary_sessions_test.go`

**Interfaces:**
- Consumes: `summary.SessionSummaryModel` (Task 3); existing helper `formatCost`.
- Produces (used by Task 7): `func PrintSessionSummary(out io.Writer, m *summary.SessionSummaryModel) error`.

- [ ] **Step 1: Write the failing test**

Create `internal/output/summary_sessions_test.go`:

```go
package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

func TestPrintSessionSummary(t *testing.T) {
	m := &summary.SessionSummaryModel{
		ID: "s_abc", Agent: "driveaway-agent", TurnCount: 2,
		Duration: "2m0s", Cost: f64(0.08),
		Turns: []summary.Turn{
			{Number: 1, Customer: "I want to rent a car next weekend",
				Agent: "Great! We have 3 cars available...", Tools: []string{"check_availability"}},
			{Number: 2, Customer: "Downtown", Agent: "Booked! Confirmation #1234",
				Tools: []string{"create_booking"}, Journeys: []string{"rental"}},
		},
	}
	var buf bytes.Buffer
	if err := PrintSessionSummary(&buf, m); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"Session s_abc · driveaway-agent · 2 turns · 2m0s",
		"Turn 1",
		"Customer: I want to rent a car next weekend",
		"[tools: check_availability]",
		"Agent: Great! We have 3 cars available...",
		"Turn 2",
		"[tools: create_booking] [journey: rental]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n---\n%s", want, got)
		}
	}
}

func TestPrintSessionSummary_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintSessionSummary(&buf, &summary.SessionSummaryModel{ID: "s_x"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No turns found.") {
		t.Fatalf("expected empty-state message, got %q", buf.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/output/ -run TestPrintSessionSummary -v`
Expected: FAIL — `undefined: PrintSessionSummary`.

- [ ] **Step 3: Write `internal/output/summary_sessions.go`**

```go
package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

const sessionIndent = "        " // 8 spaces, aligns under "Turn N  "

// PrintSessionSummary renders a conversation overview: transcript + event tags.
func PrintSessionSummary(out io.Writer, m *summary.SessionSummaryModel) error {
	var b strings.Builder

	header := "Session " + m.ID
	if m.Agent != "" {
		header += " · " + m.Agent
	}
	header += fmt.Sprintf(" · %d turns", m.TurnCount)
	if m.Duration != "" {
		header += " · " + m.Duration
	}
	if m.Cost != nil {
		header += " · " + formatCost(m.Cost)
	}
	b.WriteString(header + "\n\n")

	if len(m.Turns) == 0 {
		b.WriteString("No turns found.\n")
		_, err := io.WriteString(out, b.String())
		return err
	}

	for _, turn := range m.Turns {
		b.WriteString(fmt.Sprintf("Turn %-3d Customer: %s\n", turn.Number, turn.Customer))

		var tags []string
		if len(turn.Tools) > 0 {
			tags = append(tags, "[tools: "+strings.Join(turn.Tools, ", ")+"]")
		}
		if len(turn.Journeys) > 0 {
			tags = append(tags, "[journey: "+strings.Join(turn.Journeys, ", ")+"]")
		}
		if len(tags) > 0 {
			b.WriteString(sessionIndent + strings.Join(tags, " ") + "\n")
		}

		if turn.Agent != "" {
			b.WriteString(sessionIndent + "Agent: " + turn.Agent + "\n")
		}
	}

	_, err := io.WriteString(out, b.String())
	return err
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/output/ -run TestPrintSessionSummary -v`
Expected: PASS. (Note: `fmt.Sprintf("Turn %-3d Customer:", 1)` → `"Turn 1   Customer:"`; the test asserts the substring `"Turn 1"` and `"Customer: ..."`, both present.)

- [ ] **Step 5: Commit**

```bash
git add internal/output/summary_sessions.go internal/output/summary_sessions_test.go
git commit -m "feat(output): render session conversation overview"
```

---

### Task 6: Wire `--summary` into `traces get`

**Files:**
- Modify: `cmd/traces.go` (var block ~44-49; `tracesGetCmd` RunE ~189-214; `init()` ~347-356)

**Interfaces:**
- Consumes: `apiClient.GetTrace(ctx, org, proj, traceID, fields) (*clients.TraceDetail, json.RawMessage, error)`; `apiClient.ListObservations(ctx, org, proj, traceID, includeIO bool) ([]clients.ObservationInfo, json.RawMessage, error)`; `summary.TraceSummary`; `output.PrintTraceSummary`.
- Produces: the `iai traces get <id> --summary` user-facing behavior.

- [ ] **Step 1: Add the flag variable**

In `cmd/traces.go`, add to the `var (...)` block (near `tracesGetFields` ~line 43):

```go
	tracesGetSummary bool
```

- [ ] **Step 2: Add the summary branch in `tracesGetCmd.RunE`**

In `cmd/traces.go`, add the import for the summary package at the top:

```go
	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
```

Then in `tracesGetCmd.RunE`, immediately after `resolveProject(...)` returns successfully and BEFORE the existing `apiClient.GetTrace(...)` call, insert:

```go
		if tracesGetSummary {
			trace, _, err := apiClient.GetTrace(
				cmd.Context(), pCtx.orgId, pCtx.projectId, traceID, "core,io,metrics",
			)
			if err != nil {
				return err
			}
			obs, _, err := apiClient.ListObservations(
				cmd.Context(), pCtx.orgId, pCtx.projectId, traceID, true,
			)
			if err != nil {
				return err
			}
			return output.PrintTraceSummary(out, summary.TraceSummary(trace, obs))
		}
```

(Leave the existing `GetTrace`/JSON/YAML/`PrintTraceDetail` code below unchanged.)

- [ ] **Step 3: Register the flag and update mutual exclusion + example**

In `cmd/traces.go` `init()`, in the "traces get flags" section (~348-356):

```go
	tracesGetCmd.Flags().BoolVar(&tracesGetSummary, "summary", false,
		"Render a compact, LLM-readable summary of the turn (conditions, tools, iterations)")
	tracesGetCmd.MarkFlagsMutuallyExclusive("summary", "json", "yaml")
```

Replace the existing `tracesGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")` line with the three-way version above (do not keep both).

Add to `tracesGetCmd.Example` (the `Example:` string near line 185):

```
  iai traces get abc123 --summary
```

- [ ] **Step 4: Build and verify the flag exists**

Run: `go build ./... && go run main.go traces get --help`
Expected: build succeeds; help text lists `--summary` and shows the example. (No network call made by `--help`.)

- [ ] **Step 5: Run the full test suite (no regressions)**

Run: `go test ./...`
Expected: PASS across all packages.

- [ ] **Step 6: Commit**

```bash
git add cmd/traces.go
git commit -m "feat(traces): add --summary flag to traces get"
```

---

### Task 7: Add IO fields to `TraceInfo`; wire `--summary` into `sessions get`

**Files:**
- Modify: `internal/clients/api_client.go:357-376` (`TraceInfo`)
- Modify: `cmd/sessions.go` (var block ~22-28; `sessionsGetCmd` RunE ~127-157; `init()` ~185-194)

**Interfaces:**
- Consumes: `apiClient.ListTraces(ctx, org, proj, clients.TraceListOptions) ([]clients.TraceInfo, clients.TraceMeta, json.RawMessage, error)`; `summary.SessionSummary`; `output.PrintSessionSummary`.
- Produces: the `iai sessions get <id> --summary` behavior; `clients.TraceInfo.Input`/`.Output` fields (consumed by Task 3).

> If Task 3 was implemented first per its ordering note, Step 1 below is already done — verify and skip.

- [ ] **Step 1: Add IO fields to `TraceInfo`**

In `internal/clients/api_client.go`, in the `TraceInfo` struct (lines 357-376), add after the `Level string` field:

```go
	Input  json.RawMessage `json:"input,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`
```

(`encoding/json` is already imported in this file. These fields are populated only when the trace-list endpoint is queried with `fields=...,io`; the existing list table and `--json` paths are unaffected because they never reference these fields.)

- [ ] **Step 2: Verify the trace-list endpoint returns IO (contingency check)**

Run (requires auth — see note): `go run main.go traces list --session-id <known-session-id> --fields core,io --json | jq '.data.traces[0] | {input, output, tags}'`
Expected: `input` and `output` are non-null for a real turn.
- If you lack credentials, defer this check to Step 6's manual smoke test.
- **If `input`/`output` come back null** (endpoint ignores `io` on list): change the session handler in Step 3 to fetch per-turn via `GetTrace(..., "core,io")` in a loop over the session's traces instead of relying on list IO. Keep everything else identical.

- [ ] **Step 3: Add the summary branch in `sessionsGetCmd.RunE`**

In `cmd/sessions.go`, add imports:

```go
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
```

(`clients` and `time` may already be imported — do not duplicate.)

In `sessionsGetCmd.RunE`, after `resolveProject(...)` succeeds and BEFORE the existing `apiClient.GetSession(...)` call, insert:

```go
		if sessionsGetSummary {
			var all []clients.TraceInfo
			page := 1
			for {
				traces, meta, _, err := apiClient.ListTraces(
					cmd.Context(), pCtx.orgId, pCtx.projectId,
					clients.TraceListOptions{
						SessionID: sessionID,
						Fields:    "core,io",
						Order:     "asc",
						OrderBy:   "timestamp",
						Limit:     100,
						Page:      page,
					},
				)
				if err != nil {
					return err
				}
				all = append(all, traces...)
				if meta.TotalPages <= page || len(traces) == 0 {
					break
				}
				page++
			}
			return output.PrintSessionSummary(out, summary.SessionSummary(sessionID, all))
		}
```

(Leave the existing `GetSession`/JSON/YAML/`PrintSessionDetail` code below unchanged.)

- [ ] **Step 4: Add the flag variable and registration**

In `cmd/sessions.go` `var (...)` block, add:

```go
	sessionsGetSummary bool
```

In `init()`, in the "sessions get flags" section (~185-194):

```go
	sessionsGetCmd.Flags().BoolVar(&sessionsGetSummary, "summary", false,
		"Render a compact, LLM-readable overview of the conversation (transcript + event tags)")
	sessionsGetCmd.MarkFlagsMutuallyExclusive("summary", "json", "yaml")
```

Replace the existing `sessionsGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")` with the three-way version above.

Add to `sessionsGetCmd.Example`:

```
  iai sessions get <session-id> --summary
```

- [ ] **Step 5: Build and run the full test suite**

Run: `go build ./... && go test ./...`
Expected: build succeeds; all tests pass (including `internal/summary` session tests that depend on the new `TraceInfo` fields).

- [ ] **Step 6: Manual smoke test (if credentials available)**

Run (requires `iai login` or `INTERACTIVE_API_KEY`):
```
go run main.go traces get <real-trace-id> --summary
go run main.go sessions get <real-session-id> --summary
```
Expected: trace shows per-iteration conditions + tools + agent reply; session shows transcript + `[tools: ...]`/`[journey: ...]` tags. If field shapes differ from the assumptions in Task 2/3 (e.g. envelope or IO encoding), adjust the parsing helpers and re-run the unit tests. If no credentials, note this in the PR as "manual smoke pending."

- [ ] **Step 7: Commit**

```bash
git add internal/clients/api_client.go cmd/sessions.go
git commit -m "feat(sessions): add --summary flag and TraceInfo IO fields"
```

---

### Task 8: Regenerate command docs

**Files:**
- Modify: `docs/iai_traces_get.md`, `docs/iai_sessions_get.md` (regenerated, not hand-edited)

**Interfaces:**
- Consumes: the cobra command definitions updated in Tasks 6–7.
- Produces: updated reference docs reflecting `--summary`.

- [ ] **Step 1: Regenerate docs**

Run: `make docs`
Expected: `go run main.go gen-docs` runs; `git status` shows `docs/iai_traces_get.md` and `docs/iai_sessions_get.md` modified to include the `--summary` flag and new examples.

- [ ] **Step 2: Verify no unintended doc churn**

Run: `git status --short docs/`
Expected: only `iai_traces_get.md` and `iai_sessions_get.md` changed. If other docs changed (e.g. a global timestamp), inspect the diff; revert any unrelated noise.

- [ ] **Step 3: Commit**

```bash
git add docs/
git commit -m "docs: regenerate reference for traces/sessions --summary"
```

---

## Self-Review

**Spec coverage:**
- `--summary` on `traces get` → Task 6. ✅
- `--summary` on `sessions get` → Task 7. ✅
- Mutually exclusive with `--json`/`--yaml` → Tasks 6 & 7 Step (MarkFlagsMutuallyExclusive). ✅
- Trace view: per-iteration conditions + tools + reply + errors → Tasks 2 & 4. ✅
- Session view: transcript + event tags (tools/journey), header → Tasks 3 & 5. ✅
- No client signature changes except `TraceInfo` IO fields → Task 7 Step 1. ✅
- Token discipline (500/160 truncation, rationale omitted, rune-safe) → Task 1 (`Truncate`, `MaxValueLen`, `MaxSessionMsgLen`); rationale never extracted in Task 2. ✅
- JSON-string-wrapped IO handling → Task 1 (`UnwrapJSON`/`AsString`). ✅
- Robust to missing/unknown spans → Task 2 (skips unrecognized; empty-obs path tested). ✅
- Docs regenerated → Task 8. ✅
- Existing output untouched → Tasks 6/7 insert branches before existing code; no deletions. ✅

**Placeholder scan:** No TBD/TODO; every code step contains complete code; commands have expected output. ✅

**Type consistency:** `TraceSummary(*clients.TraceDetail, []clients.ObservationInfo) *TraceSummaryModel`, `SessionSummary(string, []clients.TraceInfo) *SessionSummaryModel`, `PrintTraceSummary(io.Writer, *summary.TraceSummaryModel)`, `PrintSessionSummary(io.Writer, *summary.SessionSummaryModel)` — names/signatures match across producing and consuming tasks. Helper names (`AsString`, `UnwrapJSON`, `CompactArgs`, `CompactJSON`, `Truncate`, `CollapseWS`) consistent between Task 1 definitions and Task 2/3 uses. ✅

**Known cross-task ordering:** Task 3 depends on `TraceInfo` IO fields defined in Task 7 Step 1 — flagged in Task 3 with instructions to perform that edit first. If implementing strictly in number order, do Task 7 Step 1 before Task 3 Step 2.
