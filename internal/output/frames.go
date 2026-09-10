package output

import (
	"fmt"
	"io"
	"strings"
)

// Redraw prepares a frame to overwrite the one before it: every line clears to
// its end and the frame clears whatever followed it, so a shrinking frame
// leaves nothing behind and there is no blank flash between refreshes.
func Redraw(frame string) string {
	return strings.ReplaceAll(frame, "\n", "\033[K\n") + "\033[J"
}

// FrameWriter replaces its previous frame in place on a terminal, and appends
// frames otherwise, for progress that updates while a command runs.
type FrameWriter struct {
	w     io.Writer
	tty   bool
	lines int
}

func NewFrameWriter(w io.Writer) *FrameWriter {
	return &FrameWriter{w: w, tty: IsTerminal(w)}
}

// Write replaces the previous frame. The frame must end with a newline.
func (f *FrameWriter) Write(frame string) {
	if !f.tty {
		fmt.Fprint(f.w, frame)
		return
	}
	if f.lines > 0 {
		fmt.Fprintf(f.w, "\033[%dA", f.lines)
	}
	fmt.Fprint(f.w, Redraw(frame))
	f.lines = strings.Count(frame, "\n")
}
