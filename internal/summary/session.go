package summary

import (
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
	var agents []string
	agentSeen := map[string]bool{}
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
			case strings.HasPrefix(tag, "agent:"):
				// A session can carry more than one agent (e.g. a shadow/dev
				// deployment logging alongside production); surface them all.
				if a := strings.TrimPrefix(tag, "agent:"); a != "" && !agentSeen[a] {
					agentSeen[a] = true
					agents = append(agents, a)
				}
			}
		}
		m.Turns = append(m.Turns, turn)
	}

	m.Agent = strings.Join(agents, ", ")
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
	// RFC3339Nano also parses plain RFC3339 — the fractional seconds are optional.
	return time.Parse(time.RFC3339Nano, s)
}
