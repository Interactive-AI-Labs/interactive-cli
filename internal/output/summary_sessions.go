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
