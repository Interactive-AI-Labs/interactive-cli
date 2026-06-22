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
					// Guideline text can carry raw newlines and run very long;
					// normalize whitespace so it dedupes and reads on one line.
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
			case spanExecuteToolCalls:
				// Tools are the direct children of execute_tool_calls.
				for _, tool := range children[d.ID] {
					tc := ToolCall{
						Name:   tool.Name,
						Args:   Truncate(CompactArgs(tool.Input), MaxValueLen),
						Result: Truncate(CompactJSON(UnwrapToolResult(tool.Output)), MaxValueLen),
					}
					if strings.EqualFold(tool.Level, "ERROR") {
						tc.Errored = true
						tc.ErrMsg = tool.StatusMessage
					}
					iteration.Tools = append(iteration.Tools, tc)
				}
			}
		}

		for _, c := range condOrder {
			iteration.Conditions = append(iteration.Conditions, Condition{Text: Truncate(c, MaxValueLen), Score: condScore[c]})
		}

		m.Iterations = append(m.Iterations, iteration)
	}

	m.KB = knowledgeBase(obs)
	return m
}

// knowledgeBase summarizes the turn's knowledge-base retrievals across the whole
// observation tree (not just inside iterations): the engine emits the titled
// retriever:knowledge_base span at the root, while per-iteration
// find_similar_documents spans carry only untitled content.
func knowledgeBase(obs []clients.ObservationInfo) *KBRetrieval {
	var titles []string
	seen := map[string]bool{}
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
			t = Truncate(CollapseWS(t), MaxKBTitleLen)
			if t != "" && !seen[t] {
				seen[t] = true
				titles = append(titles, t)
			}
		}
		if count > curatedCount {
			curatedCount = count
		}
	}
	if !used {
		return nil
	}
	// The curated retriever result is the high-signal "which docs came back";
	// untitled find_similar plumbing only sets a count when nothing curated ran.
	kb := &KBRetrieval{Docs: titles}
	if len(titles) > 0 {
		kb.Count = curatedCount
		if kb.Count < len(titles) {
			kb.Count = len(titles)
		}
	} else {
		kb.Count = rawMax
	}
	return kb
}

// kbDocsFromOutput extracts retrieved-document titles and a document count from a
// KB span's output, handling both the retriever:knowledge_base object shape
// ({"articles":[{"name":…}],"article_count":N}) and the find_similar_documents
// array shape ([{…}], usually untitled).
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

	// find_similar_documents shape: a raw array of vector-search hits. These are
	// plumbing (technical doc ids/synonyms), not the curated result — count only.
	var arr []json.RawMessage
	if json.Unmarshal(out, &arr) == nil {
		return nil, len(arr)
	}

	return nil, 0
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
