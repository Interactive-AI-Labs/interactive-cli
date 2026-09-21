package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// TestProgressBar_LongUnicodeLabelKeepsStatusInsideWidth catches a draw that
// trims the frame by bytes: the label eats the line and a rune is split.
func TestProgressBar_LongUnicodeLabelKeepsStatusInsideWidth(t *testing.T) {
	const width = 56
	var buf bytes.Buffer

	bar := &ProgressBar{
		out:   &buf,
		label: "Uploading 2026-отчёт-за-третий-квартал.pdf",
		total: 4 << 30,
		start: time.Now().Add(-time.Second),
		isTTY: true,
		drawn: true,
		width: func() int { return width },
	}

	bar.Add(1 << 30)
	frame := buf.String()

	if !utf8.ValidString(frame) {
		t.Errorf("frame = %q, want valid UTF-8; trimming by bytes split a rune", frame)
	}

	visible := replayLine(t, frame)
	if got := utf8.RuneCountInString(visible); got > width {
		t.Errorf("visible frame is %d runes wide, want at most %d: %q", got, width, visible)
	}
	if want := HumanBytes(1 << 30); !strings.Contains(visible, want) {
		t.Errorf("frame = %q, want it to keep the transferred bytes %q", visible, want)
	}
	if !strings.Contains(visible, "(25%)") {
		t.Errorf("frame = %q, want it to keep the percentage", visible)
	}
	if !rateField.MatchString(visible) {
		t.Errorf("frame = %q, want it to keep the rate", visible)
	}
}

func TestProgressBar_UsesCurrentTerminalWidthForEachDraw(t *testing.T) {
	var buf bytes.Buffer
	width := 48
	label := "Downloading a-file-with-a-long-name.bin"
	bar := &ProgressBar{
		out:   &buf,
		label: label,
		total: 4 << 30,
		start: time.Now().Add(-time.Second),
		isTTY: true,
		drawn: true,
		width: func() int { return width },
	}

	bar.Add(1 << 30)
	firstFrame := buf.String()
	if visible := replayLine(t, firstFrame); strings.Contains(visible, label) {
		t.Fatalf("narrow frame = %q, want the label trimmed", visible)
	}

	width = 120
	bar.Add(1 << 30)
	secondFrame := buf.String()[len(firstFrame):]
	if visible := replayLine(t, secondFrame); !strings.Contains(visible, label) {
		t.Errorf("wide frame = %q, want the complete label %q", visible, label)
	}
}
