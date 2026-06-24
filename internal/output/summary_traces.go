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

	var latency, cost, errTag string
	ts := LocalTime(m.Timestamp) // "" for an empty/unparseable timestamp
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

	if input := truncateValue(m.Input, maxValueLen); input != "" {
		b.WriteString("Customer: " + input + "\n\n")
	}

	if m.KB != nil {
		b.WriteString("Knowledge base: " + formatKB(m.KB) + "\n\n")
	}

	for _, it := range m.Iterations {
		b.WriteString(fmt.Sprintf("Iteration %d\n", it.Number))
		if len(it.Routines) > 0 {
			b.WriteString("  Routines: " + strings.Join(it.Routines, ", ") + "\n")
		}
		if len(it.Journey) > 0 {
			b.WriteString("  Journey:\n")
			for _, j := range it.Journey {
				b.WriteString(fmt.Sprintf("    ▸ %s ▸ %s\n", j.Routine, j.Step))
				if cond := truncateValue(j.Condition, maxValueLen); cond != "" {
					b.WriteString("        (" + cond + ")\n")
				}
			}
		}
		for _, d := range it.Decisions {
			b.WriteString("  Why: " + truncateValue(d, maxValueLen) + "\n")
		}
		if len(it.Conditions) > 0 {
			b.WriteString("  Conditions met:\n")
			for _, c := range it.Conditions {
				b.WriteString(
					fmt.Sprintf("    ✓ %s (%d)\n", truncateValue(c.Text, maxValueLen), c.Score),
				)
			}
		}
		if len(it.Tools) > 0 {
			b.WriteString("  Tools called:\n")
			for _, tc := range it.Tools {
				args := truncateValue(compactArgs(tc.Args), maxValueLen)
				line := fmt.Sprintf("    → %s(%s)", tc.Name, args)
				if tc.Errored {
					line += " → ERROR: " + tc.ErrMsg
				} else if result := truncateValue(summary.CompactJSON(tc.Result), maxValueLen); result != "" {
					line += " → " + result
				}
				b.WriteString(line + "\n")
			}
		} else {
			b.WriteString("  (no tools called)\n")
		}
	}

	if reply := truncateValue(m.Reply, maxValueLen); reply != "" {
		b.WriteString("\nAgent: " + reply + "\n")
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

// formatKB renders a turn's knowledge-base retrieval: the document count, plus
// the retrieved titles when the backend supplied them.
func formatKB(kb *summary.KBRetrieval) string {
	noun := "docs"
	if kb.Count == 1 {
		noun = "doc"
	}
	if len(kb.Docs) > 0 {
		quoted := make([]string, len(kb.Docs))
		for i, d := range kb.Docs {
			quoted[i] = fmt.Sprintf("%q", truncateValue(d, maxKBTitleLen))
		}
		return fmt.Sprintf("%d %s retrieved — %s", kb.Count, noun, TruncateList(quoted, 6))
	}
	return fmt.Sprintf("%d %s retrieved", kb.Count, noun)
}
