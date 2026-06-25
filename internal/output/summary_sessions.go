package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

// PrintSessionSummaryTruncationWarning warns that the session summary is built
// from a capped number of turns, so trailing turns may be missing.
func PrintSessionSummaryTruncationWarning(errOut io.Writer, maxTraces int) {
	printWarning(
		errOut,
		fmt.Sprintf(
			"Warning: session summary is truncated at %d turns; later turns are omitted.",
			maxTraces,
		),
		false,
	)
}

// PrintSessionSummary renders a conversation overview: transcript + event tags.
func PrintSessionSummary(out io.Writer, m *summary.SessionSummaryModel) error {
	var b strings.Builder

	turnNoun := "turns"
	if m.TurnCount == 1 {
		turnNoun = "turn"
	}
	cost := ""
	if m.Cost != nil {
		cost = formatCost(m.Cost)
	}
	header := joinHeader(
		"Session "+m.ID,
		strings.Join(m.Agents, ", "),
		fmt.Sprintf("%d %s", m.TurnCount, turnNoun),
		m.Duration,
		cost,
	)
	b.WriteString(header + "\n\n")

	if len(m.Turns) == 0 {
		b.WriteString("No turns found.\n")
	}

	for _, turn := range m.Turns {
		fmt.Fprintf(&b, "Turn %d\n", turn.Number)
		b.WriteString("  Customer: " + truncateValue(turn.Customer, maxSessionMsgLen) + "\n")

		var tags []string
		if len(turn.Tools) > 0 {
			tags = append(tags, "[tools: "+strings.Join(turn.Tools, ", ")+"]")
		}
		if len(turn.Journeys) > 0 {
			tags = append(tags, "[journey: "+strings.Join(turn.Journeys, ", ")+"]")
		}
		if len(tags) > 0 {
			b.WriteString("  " + strings.Join(tags, " ") + "\n")
		}

		if agent := truncateValue(turn.Agent, maxSessionMsgLen); agent != "" {
			b.WriteString("  Agent: " + agent + "\n")
		}
	}

	_, err := io.WriteString(out, b.String())
	return err
}
