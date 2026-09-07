package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"
)

// progressDrawDelay keeps small/fast transfers from drawing a bar at all.
const progressDrawDelay = 500 * time.Millisecond

// ProgressBar renders an in-place terminal progress bar; a no-op off a TTY.
type ProgressBar struct {
	out     io.Writer
	label   string
	total   int64
	current int64
	start   time.Time
	drawn   bool
	isTTY   bool
}

// NewProgressBar creates a bar for a transfer of total bytes, labeled label.
func NewProgressBar(out io.Writer, total int64, label string) *ProgressBar {
	return &ProgressBar{
		out:   out,
		label: label,
		total: total,
		start: time.Now(),
		isTTY: IsTerminal(out),
	}
}

// Add advances the bar by n bytes and redraws if appropriate.
func (p *ProgressBar) Add(n int64) {
	p.current += n
	if !p.isTTY {
		return
	}
	if !p.drawn && time.Since(p.start) < progressDrawDelay {
		return
	}
	p.drawn = true
	p.draw()
}

// Finish redraws once more and moves to a new line, if the bar ever drew.
func (p *ProgressBar) Finish() {
	if !p.isTTY || !p.drawn {
		return
	}
	p.draw()
	fmt.Fprintln(p.out)
}

func (p *ProgressBar) draw() {
	elapsed := time.Since(p.start).Seconds()
	rate := float64(0)
	if elapsed > 0 {
		rate = float64(p.current) / elapsed
	}

	pct := float64(0)
	if p.total > 0 {
		pct = float64(p.current) / float64(p.total) * 100
	}

	line := fmt.Sprintf(
		"\r%s: %s / %s (%.0f%%) %s/s",
		p.label,
		HumanBytes(p.current),
		HumanBytes(p.total),
		pct,
		HumanBytes(int64(rate)),
	)

	if width, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && width > 0 && len(line) > width {
		line = line[:width]
	}

	fmt.Fprint(p.out, line)
}
