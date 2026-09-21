package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// progressDrawDelay keeps small/fast transfers from drawing a bar at all.
const progressDrawDelay = 500 * time.Millisecond

// ProgressBar renders an in-place terminal progress bar; a no-op off a TTY.
type ProgressBar struct {
	mu       sync.Mutex
	out      io.Writer
	label    string
	total    int64
	current  int64
	start    time.Time
	drawn    bool
	finished bool
	isTTY    bool
	width    func() int
}

func NewProgressBar(out io.Writer, total int64, label string) *ProgressBar {
	return &ProgressBar{
		out:   out,
		label: label,
		total: total,
		start: time.Now(),
		isTTY: IsTerminal(out),
		width: func() int { return terminalWidth(out) },
	}
}

// SetTotal exists because a download's total is only known once the response arrives.
func (p *ProgressBar) SetTotal(n int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}
	p.total = n
}

func (p *ProgressBar) Add(n int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}
	p.current += n
	if !p.isTTY {
		return
	}
	if !p.drawn && time.Since(p.start) < progressDrawDelay {
		return
	}
	p.drawn = true
	p.drawLocked()
}

func (p *ProgressBar) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}
	p.finished = true
	if !p.isTTY || !p.drawn {
		return
	}
	p.drawLocked()
	fmt.Fprintln(p.out)
}

func (p *ProgressBar) draw() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.drawLocked()
}

func (p *ProgressBar) drawLocked() {
	elapsed := time.Since(p.start).Seconds()
	rate := float64(0)
	if elapsed > 0 {
		rate = float64(p.current) / elapsed
	}

	status := fmt.Sprintf(": %s %s/s", HumanBytes(p.current), HumanBytes(int64(rate)))
	if p.total > 0 {
		status = fmt.Sprintf(
			": %s / %s (%.0f%%) %s/s",
			HumanBytes(p.current),
			HumanBytes(p.total),
			float64(p.current)/float64(p.total)*100,
			HumanBytes(int64(rate)),
		)
	}

	// Trim the label, never the status: a long filename must not cost the numbers.
	label := p.label
	width := 0
	if p.width != nil {
		width = p.width()
	}
	if width > 0 {
		if budget := width - 1 - utf8.RuneCountInString(
			status,
		); budget < utf8.RuneCountInString(
			label,
		) {
			label = trimRunes(label, budget)
		}
	}

	fmt.Fprint(p.out, "\r"+label+status+"\x1b[K")
}

func trimRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func terminalWidth(out io.Writer) int {
	f, ok := out.(*os.File)
	if !ok {
		return 0
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return 0
	}
	return width
}
