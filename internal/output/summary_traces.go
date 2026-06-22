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

	var ts, latency, cost, errTag string
	if t := LocalTime(m.Timestamp); t != "" && m.Timestamp != "" {
		ts = t
	}
	if m.LatencyMs != nil {
		latency = formatLatencyMs(m.LatencyMs)
	}
	if m.Cost != nil {
		cost = formatCost(m.Cost)
	}
	if strings.EqualFold(m.Level, "ERROR") {
		errTag = "ERROR"
	}
	iterNoun := "iterations"
	if len(m.Iterations) == 1 {
		iterNoun = "iteration"
	}
	header := joinHeader(
		"Turn — "+m.Name,
		ts, latency, cost, errTag,
		fmt.Sprintf("%d %s", len(m.Iterations), iterNoun),
	)
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
