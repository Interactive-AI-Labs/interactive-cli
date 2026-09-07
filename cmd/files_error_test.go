package cmd

import (
	"errors"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/spf13/cobra"
)

type failingFilesErrorWriter struct {
	err error
}

func (w failingFilesErrorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestReportFileRefAmbiguous_PreservesDomainAndRenderErrors(t *testing.T) {
	writeErr := errors.New("writer closed")
	ambiguousErr := &platform.FileRefAmbiguousError{Ref: "report.pdf"}

	cmd := &cobra.Command{}
	cmd.SetErr(failingFilesErrorWriter{err: writeErr})
	err := reportFileRefAmbiguous(cmd, ambiguousErr)

	if !errors.Is(err, ambiguousErr) {
		t.Errorf("error = %v, want it to preserve the ambiguous-reference error", err)
	}
	if !errors.Is(err, writeErr) {
		t.Errorf("error = %v, want it to preserve the rendering error", err)
	}
}
