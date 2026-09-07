package output

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

var rateField = regexp.MustCompile(`[0-9]+(\.[0-9]+)? ([KMGTPE]iB|B)/s`)

// replayLine reduces emitted bytes to the single line a terminal would show,
// honouring carriage return and erase-to-end-of-line.
func replayLine(t *testing.T, emitted string) string {
	t.Helper()

	var line []rune
	col := 0
	for i := 0; i < len(emitted); {
		switch {
		case emitted[i] == '\r':
			col = 0
			i++
		case strings.HasPrefix(emitted[i:], "\x1b[K"):
			line = line[:col]
			i += 3
		case strings.HasPrefix(emitted[i:], "\x1b[0K"):
			line = line[:col]
			i += 4
		case emitted[i] == '\x1b':
			t.Fatalf("emitted an escape sequence this replay does not model: %q", emitted[i:])
		default:
			r, size := utf8.DecodeRuneInString(emitted[i:])
			if col < len(line) {
				line[col] = r
			} else {
				line = append(line, r)
			}
			col++
			i += size
		}
	}
	return strings.TrimRight(string(line), " ")
}

// TestProgressBar_ShorterFrameLeavesNoResidue catches a draw that returns to
// column 0 without clearing, so the previous frame's tail stays on screen.
func TestProgressBar_ShorterFrameLeavesNoResidue(t *testing.T) {
	var buf bytes.Buffer
	bar := &ProgressBar{
		out:   &buf,
		label: "Downloading data.bin",
		total: 4 << 30,
		start: time.Now().Add(-time.Second),
		isTTY: true,
		drawn: true,
	}

	bar.Add(640 << 20)
	first := buf.String()

	bar.start = time.Now().Add(-100 * time.Hour) // collapses the rate, shrinking the frame
	bar.Add(1)
	second := buf.String()[len(first):]

	visibleFirst := replayLine(t, first)
	visibleSecond := replayLine(t, second)
	if visibleSecond == "" {
		t.Fatalf("second frame rendered nothing: %q", second)
	}
	if utf8.RuneCountInString(visibleSecond) >= utf8.RuneCountInString(visibleFirst) {
		t.Fatalf(
			"second frame %q is not shorter than the first %q; the test no longer exercises a shrinking frame",
			visibleSecond,
			visibleFirst,
		)
	}

	if got := replayLine(t, buf.String()); got != visibleSecond {
		t.Errorf(
			"visible line = %q, want %q; the first frame's tail is still on screen",
			got,
			visibleSecond,
		)
	}
}

// TestProgressBar_UnknownTotalOmitsTotalAndPercent catches a bar that reports
// "/ 0 B (0%)" while the response has not declared a length.
func TestProgressBar_UnknownTotalOmitsTotalAndPercent(t *testing.T) {
	var buf bytes.Buffer
	bar := &ProgressBar{
		out:   &buf,
		label: "Downloading data.bin",
		start: time.Now().Add(-4 * time.Second),
		isTTY: true,
		drawn: true,
	}

	bar.Add(41 << 20)
	frame := buf.String()
	visible := replayLine(t, frame)

	if strings.Contains(visible, "/ 0 B") {
		t.Errorf("frame = %q, want no zero total when the total is unknown", visible)
	}
	if strings.Contains(visible, "(0%)") {
		t.Errorf("frame = %q, want no percentage when the total is unknown", visible)
	}
	if want := HumanBytes(41 << 20); !strings.Contains(visible, want) {
		t.Errorf("frame = %q, want it to contain the transferred bytes %q", visible, want)
	}
	if !rateField.MatchString(visible) {
		t.Errorf("frame = %q, want it to contain a rate", visible)
	}
}

// TestProgressBar_FastTransferDrawsNothing pins Finish short-circuiting for a
// transfer that ends inside progressDrawDelay.
func TestProgressBar_FastTransferDrawsNothing(t *testing.T) {
	var buf bytes.Buffer
	bar := &ProgressBar{
		out:   &buf,
		label: "Uploading small.txt",
		total: 12,
		start: time.Now(),
		isTTY: true,
	}

	bar.Add(12)
	bar.Finish()

	if got := buf.String(); got != "" {
		t.Errorf(
			"output = %q, want nothing drawn for a transfer shorter than %s",
			got,
			progressDrawDelay,
		)
	}
}
