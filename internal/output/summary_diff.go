package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

// PrintTraceDiff renders a side-by-side comparison of two turns: routine and
// tool set-diffs, then the per-iteration journey path with its divergence point.
func PrintTraceDiff(out io.Writer, m *summary.TraceDiffModel) error {
	var b strings.Builder

	b.WriteString("Trace diff\n")
	b.WriteString("  " + diffSideHeader("A", m.A) + "\n")
	b.WriteString("  " + diffSideHeader("B", m.B) + "\n")

	writeSetDiff(&b, "Routines", m.Routines)
	writeSetDiff(&b, "Tools", m.Tools)

	if len(m.Journey) > 0 {
		b.WriteString("\nJourney\n")
		for _, j := range m.Journey {
			line := fmt.Sprintf(
				"  iter %d  A: %s  B: %s",
				j.Iteration, joinLabels(j.A), joinLabels(j.B),
			)
			if j.Diverged {
				line += "  ◀ diverges"
			}
			b.WriteString(line + "\n")
		}
	}

	if m.A.Reply != "" || m.B.Reply != "" {
		b.WriteString("\nReply\n")
		writeBlock(&b, "A", m.A.Reply, "  ")
		writeBlock(&b, "B", m.B.Reply, "  ")
	}

	_, err := io.WriteString(out, b.String())
	return err
}

func diffSideHeader(side string, s summary.DiffSide) string {
	parts := []string{s.Name, fmt.Sprintf("%d iters", s.Iterations)}
	if s.Cost != nil {
		parts = append(parts, formatCost(s.Cost))
	}
	return side + " " + s.ID + "  " + joinHeader(parts...)
}

func writeSetDiff(b *strings.Builder, title string, d summary.SetDiff) {
	if len(d.Both) == 0 && len(d.AOnly) == 0 && len(d.BOnly) == 0 {
		return
	}
	b.WriteString("\n" + title + "\n")
	if len(d.Both) > 0 {
		b.WriteString("  both: " + TruncateList(d.Both, 12) + "\n")
	}
	if len(d.AOnly) > 0 {
		b.WriteString("  A only: " + TruncateList(d.AOnly, 12) + "\n")
	}
	if len(d.BOnly) > 0 {
		b.WriteString("  B only: " + TruncateList(d.BOnly, 12) + "\n")
	}
}

func joinLabels(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}
