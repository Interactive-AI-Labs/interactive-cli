// Package summary turns raw trace/observation data into compact, LLM-readable
// summary models. It performs no I/O and no rendering: the models carry full
// untruncated data, and the output layer decides how to format and truncate it.
package summary

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// Condition is a guideline whose condition matched ("marked true") in a turn.
type Condition struct {
	Text  string `json:"text"`
	Score int    `json:"score"`
}

// ToolCall is a single tool invocation within an iteration.
type ToolCall struct {
	Name    string          `json:"name"`
	Args    json.RawMessage `json:"args,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Errored bool            `json:"errored,omitempty"`
	ErrMsg  string          `json:"error,omitempty"`
}

// JourneyStep is a routine node the matcher selected this iteration — the
// reachable follow-up that won at a routine fork. The sequence of journey steps
// across iterations is the turn's decision path.
type JourneyStep struct {
	Routine   string `json:"routine"`
	Step      string `json:"step"`
	Condition string `json:"condition,omitempty"`
}

// Iteration is one engine preparation iteration within a turn.
type Iteration struct {
	Number     int           `json:"number"`
	Routines   []string      `json:"routines,omitempty"`
	Journey    []JourneyStep `json:"journey,omitempty"`
	Conditions []Condition   `json:"conditions,omitempty"`
	Decisions  []string      `json:"decisions,omitempty"`
	Tools      []ToolCall    `json:"tools,omitempty"`
}

// KBRetrieval summarizes knowledge-base documents retrieved during a turn.
type KBRetrieval struct {
	Docs  []string `json:"docs,omitempty"`
	Count int      `json:"count"`
}

// TraceSummaryModel is the compact view of a single turn (one trace).
type TraceSummaryModel struct {
	Name       string       `json:"name"`
	Timestamp  string       `json:"timestamp,omitempty"`
	LatencyMs  *float64     `json:"latency_ms,omitempty"`
	Cost       *float64     `json:"cost,omitempty"`
	Level      string       `json:"level,omitempty"`
	Input      string       `json:"input,omitempty"`
	KB         *KBRetrieval `json:"knowledge_base,omitempty"`
	Iterations []Iteration  `json:"iterations,omitempty"`
	Reply      string       `json:"reply,omitempty"`
	Errors     []string     `json:"errors,omitempty"`
}

// Observation span names the engine emits, matched when walking a turn's tree.
const (
	spanMatchGuidelines  = "match_guidelines"
	spanExecuteToolCalls = "execute_tool_calls"
	spanKBRetriever      = "retriever:knowledge_base"
	spanFindSimilarDocs  = "find_similar_documents"
	spanNextStep         = "next-step"
)

// Guideline match types the engine emits in match_guidelines output.
const (
	matchTypeRoutine     = "routine"      // a routine activation
	matchTypeRoutineNode = "routine_node" // a selected journey follow-up
)

var iterationNameRe = regexp.MustCompile(`^preparation_iteration_(\d+)$`)

// Tool results are wrapped with engine metadata siblings around data.
// Treat only known siblings as an envelope so real payload fields are preserved.
var parlantEnvelopeKeys = map[string]bool{
	"metadata":               true,
	"control":                true,
	"canned_responses":       true,
	"canned_response_fields": true,
	"guidelines":             true,
}

// UnwrapToolResult returns the engine envelope's "data" payload when the wrapper
// shape is known, and passes any other value through unchanged.
func UnwrapToolResult(raw json.RawMessage) json.RawMessage {
	inner := UnwrapJSON(raw)
	var obj map[string]json.RawMessage
	if json.Unmarshal(inner, &obj) != nil {
		return inner
	}
	data, ok := obj["data"]
	if !ok {
		return inner
	}
	for k := range obj {
		if k != "data" && !parlantEnvelopeKeys[k] {
			return inner // an unexpected sibling: not the known envelope, leave it
		}
	}
	return data
}

type matchOutput struct {
	Matches []struct {
		Type      string `json:"type"`
		RoutineID string `json:"routine_id"`
		StepID    string `json:"step_id"`
		Condition string `json:"condition"`
		Score     int    `json:"score"`
	} `json:"matches"`
}

type nextStepOutput struct {
	AppliedConditionID   string `json:"applied_condition_id"`
	NextStepRationale    string `json:"next_step_rationale"`
	CurrentStepRationale string `json:"current_step_completed_rationale"`
}

type iterNode struct {
	num int
	id  string
}

// TraceSummary builds a compact turn summary from trace observations.
func TraceSummary(trace *clients.TraceDetail, obs []clients.ObservationInfo) *TraceSummaryModel {
	children, iters, errs := indexTraceObservations(obs)
	m := &TraceSummaryModel{
		Name:      trace.Name,
		Timestamp: trace.Timestamp,
		LatencyMs: trace.LatencyMs,
		Cost:      trace.TotalCost,
		Level:     trace.Level,
		Input:     AsString(trace.Input),
		Reply:     AsString(trace.Output),
		Errors:    errs,
	}
	for _, it := range iters {
		m.Iterations = append(m.Iterations, summarizeIteration(children, it))
	}
	m.KB = knowledgeBase(obs)
	return m
}

// indexTraceObservations groups observations by parent, collects error lines, and
// returns the preparation-iteration nodes in ascending order.
func indexTraceObservations(
	obs []clients.ObservationInfo,
) (children map[string][]clients.ObservationInfo, iters []iterNode, errs []string) {
	children = make(map[string][]clients.ObservationInfo, len(obs))
	for _, o := range obs {
		children[o.ParentObservationID] = append(children[o.ParentObservationID], o)
		if strings.EqualFold(o.Level, "ERROR") {
			msg := o.StatusMessage
			if msg == "" {
				msg = "error"
			}
			errs = append(errs, o.Name+": "+msg)
		}
		if sm := iterationNameRe.FindStringSubmatch(o.Name); sm != nil {
			n, _ := strconv.Atoi(sm[1])
			iters = append(iters, iterNode{num: n, id: o.ID})
		}
	}
	sort.Slice(iters, func(i, j int) bool { return iters[i].num < iters[j].num })
	return children, iters, errs
}

// summarizeIteration builds one iteration from the observation subtree rooted at it.
func summarizeIteration(children map[string][]clients.ObservationInfo, it iterNode) Iteration {
	iteration := Iteration{Number: it.num}

	// Keep first-seen condition order and the highest score per condition.
	condScore := map[string]int{}
	var condOrder []string
	stepSeen := map[string]bool{} // journey steps keyed by (routine, step)
	var routines, decisions []string

	for _, d := range descendants(children, it.id) {
		switch d.Name {
		case spanMatchGuidelines:
			if len(d.Output) == 0 {
				continue
			}
			var mo matchOutput
			if json.Unmarshal(UnwrapJSON(d.Output), &mo) != nil {
				continue
			}
			for _, mm := range mo.Matches {
				switch mm.Type {
				case matchTypeRoutine:
					routines = append(routines, mm.RoutineID)
					continue
				case matchTypeRoutineNode:
					key := mm.RoutineID + "\x00" + mm.StepID
					if mm.StepID != "" && !stepSeen[key] {
						stepSeen[key] = true
						iteration.Journey = append(iteration.Journey, JourneyStep{
							Routine:   mm.RoutineID,
							Step:      mm.StepID,
							Condition: CollapseWS(mm.Condition),
						})
					}
					continue
				}
				// Guideline text can carry raw newlines; normalize so it dedupes cleanly.
				cond := CollapseWS(mm.Condition)
				if cond == "" {
					continue
				}
				if prev, ok := condScore[cond]; !ok || mm.Score > prev {
					if !ok {
						condOrder = append(condOrder, cond)
					}
					condScore[cond] = mm.Score
				}
			}
		case spanNextStep:
			decisions = append(decisions, decisionRationale(d.Output))
		case spanExecuteToolCalls:
			for _, tool := range children[d.ID] {
				tc := ToolCall{
					Name:   tool.Name,
					Args:   rawOrNil(UnwrapJSON(tool.Input)),
					Result: rawOrNil(UnwrapToolResult(tool.Output)),
				}
				if strings.EqualFold(tool.Level, "ERROR") {
					tc.Errored = true
					tc.ErrMsg = tool.StatusMessage
					if tc.ErrMsg == "" {
						tc.ErrMsg = "error"
					}
				}
				iteration.Tools = append(iteration.Tools, tc)
			}
		}
	}

	iteration.Routines = dedup(routines)
	iteration.Decisions = dedup(decisions)
	for _, c := range condOrder {
		iteration.Conditions = append(
			iteration.Conditions,
			Condition{Text: c, Score: condScore[c]},
		)
	}
	return iteration
}

// KB retrievals can appear as titled curated results or untitled vector hits.
// Preserve titles when present; otherwise report only the retrieved count.
func knowledgeBase(obs []clients.ObservationInfo) *KBRetrieval {
	var titles []string
	curatedCount := 0 // docs in the curated retriever result
	rawMax := 0       // largest untitled (find_similar) retrieval
	used := false
	for _, o := range obs {
		if o.Name != spanKBRetriever && o.Name != spanFindSimilarDocs {
			continue
		}
		used = true
		ts, count := kbDocsFromOutput(o.Output)
		if len(ts) == 0 {
			if count > rawMax {
				rawMax = count
			}
			continue
		}
		for _, t := range ts {
			titles = append(titles, CollapseWS(t))
		}
		if count > curatedCount {
			curatedCount = count
		}
	}
	if !used {
		return nil
	}
	titles = dedup(titles)
	// Curated retriever results identify the documents shown to the agent.
	// Use raw vector-search counts only when no titled result exists.
	kb := &KBRetrieval{Docs: titles}
	if len(titles) > 0 {
		kb.Count = max(curatedCount, len(titles))
	} else {
		kb.Count = rawMax
	}
	return kb
}

// KB spans use different JSON shapes depending on the retrieval path.
// Extract titles from curated results and counts from raw vector hits.
func kbDocsFromOutput(raw json.RawMessage) (titles []string, count int) {
	if len(raw) == 0 {
		return nil, 0
	}
	out := UnwrapJSON(raw)

	var obj struct {
		Articles []struct {
			Name string `json:"name"`
		} `json:"articles"`
		ArticleCount int `json:"article_count"`
	}
	if json.Unmarshal(out, &obj) == nil && (len(obj.Articles) > 0 || obj.ArticleCount > 0) {
		for _, a := range obj.Articles {
			if a.Name != "" {
				titles = append(titles, a.Name)
			}
		}
		count = obj.ArticleCount
		if count < len(obj.Articles) {
			count = len(obj.Articles)
		}
		return titles, count
	}

	// Raw vector-search hits do not include useful display titles.
	// Count them, but do not expose technical document payloads.
	var arr []json.RawMessage
	if json.Unmarshal(out, &arr) == nil {
		return nil, len(arr)
	}

	return nil, 0
}

// Observation parent links can be malformed or cyclic.
// Track seen nodes so summary generation always terminates.
func descendants(
	children map[string][]clients.ObservationInfo,
	id string,
) []clients.ObservationInfo {
	var out []clients.ObservationInfo
	seen := map[string]bool{}
	stack := []string{id}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		for _, c := range children[cur] {
			out = append(out, c)
			stack = append(stack, c.ID)
		}
	}
	return out
}

// decisionRationale returns the rationale when a real journey transition fired.
// applied_condition_id "", "0", or "None" mean none did, so there's no decision.
func decisionRationale(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var ns nextStepOutput
	if json.Unmarshal(UnwrapJSON(raw), &ns) != nil {
		return ""
	}
	switch ns.AppliedConditionID {
	case "", "0", "None":
		return ""
	}
	if r := CollapseWS(ns.NextStepRationale); r != "" {
		return r
	}
	return CollapseWS(ns.CurrentStepRationale)
}

// rawOrNil normalizes an empty raw message to nil so it omits from JSON output.
func rawOrNil(r json.RawMessage) json.RawMessage {
	if len(r) == 0 {
		return nil
	}
	return r
}
