package summary

import (
	"maps"
	"slices"
)

// SetDiff partitions two label sets into shared and side-exclusive members.
type SetDiff struct {
	Both  []string `json:"both,omitempty"`
	AOnly []string `json:"a_only,omitempty"`
	BOnly []string `json:"b_only,omitempty"`
}

// JourneyDiff compares the journey follow-ups selected in one iteration across
// two turns. Diverged is true when the two sides took different routine nodes —
// the fork where the agents' decision paths split.
type JourneyDiff struct {
	Iteration int      `json:"iteration"`
	A         []string `json:"a,omitempty"`
	B         []string `json:"b,omitempty"`
	Diverged  bool     `json:"diverged,omitempty"`
}

// DiffSide is the per-turn header shown on each side of a diff.
type DiffSide struct {
	ID         string   `json:"id"`
	Name       string   `json:"name,omitempty"`
	Iterations int      `json:"iterations"`
	Cost       *float64 `json:"cost,omitempty"`
	Reply      string   `json:"reply,omitempty"`
}

// TraceDiffModel compares two turn summaries: routine activations, tools, and
// the per-iteration journey decision path.
type TraceDiffModel struct {
	A        DiffSide      `json:"a"`
	B        DiffSide      `json:"b"`
	Routines SetDiff       `json:"routines"`
	Tools    SetDiff       `json:"tools"`
	Journey  []JourneyDiff `json:"journey,omitempty"`
}

// TraceDiff compares two trace summaries side by side.
func TraceDiff(idA string, a *TraceSummaryModel, idB string, b *TraceSummaryModel) *TraceDiffModel {
	return &TraceDiffModel{
		A:        diffSide(idA, a),
		B:        diffSide(idB, b),
		Routines: setDiff(traceRoutines(a), traceRoutines(b)),
		Tools:    setDiff(traceTools(a), traceTools(b)),
		Journey:  journeyDiff(a, b),
	}
}

func diffSide(id string, m *TraceSummaryModel) DiffSide {
	return DiffSide{
		ID:         id,
		Name:       m.Name,
		Iterations: len(m.Iterations),
		Cost:       m.Cost,
		Reply:      m.Reply,
	}
}

// traceRoutines is the ordered, deduped union of routine activations in a turn.
func traceRoutines(m *TraceSummaryModel) []string {
	var all []string
	for _, it := range m.Iterations {
		all = append(all, it.Routines...)
	}
	return dedup(all)
}

// traceTools is the ordered, deduped union of tool names called in a turn.
func traceTools(m *TraceSummaryModel) []string {
	var all []string
	for _, it := range m.Iterations {
		for _, t := range it.Tools {
			all = append(all, t.Name)
		}
	}
	return dedup(all)
}

// dedup returns xs with empty strings and later duplicates removed, order kept.
func dedup(xs []string) []string {
	var out []string
	seen := make(map[string]bool, len(xs))
	for _, x := range xs {
		if x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

func setDiff(a, b []string) SetDiff {
	aset, bset := toSet(a), toSet(b)
	var d SetDiff
	for _, x := range a {
		if bset[x] {
			d.Both = append(d.Both, x)
		} else {
			d.AOnly = append(d.AOnly, x)
		}
	}
	for _, x := range b {
		if !aset[x] {
			d.BOnly = append(d.BOnly, x)
		}
	}
	slices.Sort(d.Both)
	slices.Sort(d.AOnly)
	slices.Sort(d.BOnly)
	return d
}

func journeyDiff(a, b *TraceSummaryModel) []JourneyDiff {
	aByNum, bByNum := journeyLabelsByIter(a), journeyLabelsByIter(b)
	nums := map[int]bool{}
	for n := range aByNum {
		nums[n] = true
	}
	for n := range bByNum {
		nums[n] = true
	}
	ordered := slices.Sorted(maps.Keys(nums))

	out := make([]JourneyDiff, 0, len(ordered))
	for _, n := range ordered {
		al, bl := aByNum[n], bByNum[n]
		out = append(out, JourneyDiff{
			Iteration: n,
			A:         al,
			B:         bl,
			Diverged:  !equalStringSets(al, bl),
		})
	}
	return out
}

// journeyLabelsByIter maps each iteration number to its "routine/step" labels,
// skipping iterations with no journey steps.
func journeyLabelsByIter(m *TraceSummaryModel) map[int][]string {
	out := map[int][]string{}
	for _, it := range m.Iterations {
		var labels []string
		for _, j := range it.Journey {
			labels = append(labels, j.Routine+"/"+j.Step)
		}
		if len(labels) > 0 {
			out[it.Number] = labels
		}
	}
	return out
}

func toSet(xs []string) map[string]bool {
	s := make(map[string]bool, len(xs))
	for _, x := range xs {
		s[x] = true
	}
	return s
}

func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	bset := toSet(b)
	for _, x := range a {
		if !bset[x] {
			return false
		}
	}
	return true
}
