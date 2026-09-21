package output

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestProgressBar_NonFileWriterNeverTrimsLabel keeps direct construction safe:
// without an injected width callback, a long label reaches the output untouched.
func TestProgressBar_NonFileWriterNeverTrimsLabel(t *testing.T) {
	label := strings.Repeat("x", 500)

	tests := []struct {
		name string
		out  func() (io.Writer, func() string)
	}{
		{name: "bytes.Buffer", out: func() (io.Writer, func() string) {
			buf := &bytes.Buffer{}
			return buf, buf.String
		}},
		{name: "strings.Builder", out: func() (io.Writer, func() string) {
			var b strings.Builder
			return &b, b.String
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, read := tt.out()
			bar := &ProgressBar{
				out:   out,
				label: label,
				start: time.Now(),
				isTTY: true,
			}
			bar.draw()
			if !strings.Contains(read(), label) {
				t.Errorf(
					"drawn output for a non-*os.File writer trimmed the label, want it untouched",
				)
			}
		})
	}
}

// TestProgressBar_SetTotal catches SetTotal not existing or not updating p.total,
// which is what leaves the download progress bar permanently reporting "/ 0 B".
func TestProgressBar_SetTotal(t *testing.T) {
	var buf bytes.Buffer
	bar := &ProgressBar{
		out:   &buf,
		label: "download",
		start: time.Now(),
		isTTY: true,
		drawn: true,
	}

	bar.SetTotal(2048)
	bar.Add(512)
	bar.Add(512)

	line := buf.String()
	if !strings.Contains(line, HumanBytes(2048)) {
		t.Errorf("drawn line = %q, want it to contain the new total %q", line, HumanBytes(2048))
	}
	if strings.Contains(line, "(0%)") {
		t.Errorf("drawn line = %q, still shows the 0%% placeholder despite SetTotal", line)
	}
}

func TestProgressBar_ConcurrentAddAndFinishRejectsLateUpdates(t *testing.T) {
	var buf bytes.Buffer
	enteredDraw := make(chan struct{})
	releaseDraw := make(chan struct{})
	var blockFirstDraw sync.Once
	bar := &ProgressBar{
		out:   &buf,
		label: "download",
		total: 10,
		start: time.Now().Add(-time.Second),
		isTTY: true,
		drawn: true,
		width: func() int {
			blockFirstDraw.Do(func() {
				close(enteredDraw)
				<-releaseDraw
			})
			return 80
		},
	}

	addDone := make(chan struct{})
	go func() {
		defer close(addDone)
		bar.Add(1)
	}()
	<-enteredDraw

	finishStarted := make(chan struct{})
	finishDone := make(chan struct{})
	go func() {
		close(finishStarted)
		bar.Finish()
		close(finishDone)
	}()
	<-finishStarted
	close(releaseDraw)
	<-addDone
	<-finishDone

	finishedOutput := buf.String()
	bar.Add(2)
	bar.SetTotal(20)

	if got := buf.String(); got != finishedOutput {
		t.Errorf("output after Finish = %q, want no change from %q", got, finishedOutput)
	}
	if bar.current != 1 {
		t.Errorf("current after late Add = %d, want 1", bar.current)
	}
	if bar.total != 10 {
		t.Errorf("total after late SetTotal = %d, want 10", bar.total)
	}
}
