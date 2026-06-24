package summary

import (
	"sort"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// Turn is one back-and-forth within a session.
type Turn struct {
	Number   int      `json:"number"`
	Customer string   `json:"customer,omitempty"`
	Agent    string   `json:"agent,omitempty"`
	Tools    []string `json:"tools,omitempty"`
	Journeys []string `json:"journeys,omitempty"`
}

// SessionSummaryModel is the overall view of a conversation (one session).
type SessionSummaryModel struct {
	ID        string   `json:"id"`
	Agents    []string `json:"agents,omitempty"`
	TurnCount int      `json:"turn_count"`
	Duration  string   `json:"duration,omitempty"`
	Cost      *float64 `json:"cost,omitempty"`
	Turns     []Turn   `json:"turns,omitempty"`
}

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
			Customer: CollapseWS(AsString(tr.Input)),
			Agent:    CollapseWS(AsString(tr.Output)),
		}
		for _, tag := range tr.Tags {
			switch {
			case strings.HasPrefix(tag, "tool:"):
				turn.Tools = append(turn.Tools, strings.TrimPrefix(tag, "tool:"))
			case strings.HasPrefix(tag, "routine:"):
				turn.Journeys = append(turn.Journeys, strings.TrimPrefix(tag, "routine:"))
			case strings.HasPrefix(tag, "agent:"):
				// Session traces can include multiple agent tags, such as prod
				// and shadow/dev. Surface all distinct agents instead of picking
				// one arbitrarily.
				if a := strings.TrimPrefix(tag, "agent:"); a != "" && !agentSeen[a] {
					agentSeen[a] = true
					agents = append(agents, a)
				}
			}
		}
		m.Turns = append(m.Turns, turn)
	}

	m.Agents = agents
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
	first, err1 := time.Parse(time.RFC3339Nano, sorted[0].Timestamp)
	last, err2 := time.Parse(time.RFC3339Nano, sorted[len(sorted)-1].Timestamp)
	if err1 != nil || err2 != nil || last.Before(first) {
		return ""
	}
	d := last.Sub(first).Round(time.Second)
	return d.String()
}
