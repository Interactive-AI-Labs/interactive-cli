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

// Observation span names the engine emits, matched when walking a turn's tree.
const (
	spanMatchGuidelines  = "match_guidelines"
	spanExecuteToolCalls = "execute_tool_calls"
	spanKBRetriever      = "retriever:knowledge_base"
	spanFindSimilarDocs  = "find_similar_documents"
)

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

	children := make(map[string][]clients.ObservationInfo, len(obs))
	for _, o := range obs {
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

		// Conditions are aggregated across all match_guidelines spans in the
		// subtree and deduped by text (highest score wins, first-seen order).
		condScore := map[string]int{}
		var condOrder []string

		// One pass over the iteration subtree, dispatching by span name —
		// conditions, tool calls, and KB queries are mutually exclusive.
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
			case spanExecuteToolCalls:
				// Tools are the direct children of execute_tool_calls.
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
			case spanKBRetriever, spanFindSimilarDocs:
				if q := strings.TrimSpace(AsString(d.Input)); q != "" {
					iteration.KBQueries = append(iteration.KBQueries, Truncate(CollapseWS(q), MaxValueLen))
				}
			}
		}

		for _, c := range condOrder {
			iteration.Conditions = append(iteration.Conditions, Condition{Text: c, Score: condScore[c]})
		}

		m.Iterations = append(m.Iterations, iteration)
	}

	return m
}

// descendants returns all transitive children of id (excluding id itself).
func descendants(children map[string][]clients.ObservationInfo, id string) []clients.ObservationInfo {
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
