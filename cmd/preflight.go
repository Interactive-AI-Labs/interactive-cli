package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/preflight"
)

// runUpdatePreflight prints the deploy-awareness banner for a pending
// agent/service update — or the fail-open note when live state could not be
// fetched — and enforces --expect-revision when set. Everything goes to errW
// (stderr) so stdout and exit codes stay identical for scripts; the only new
// error path is the opt-in --expect-revision mismatch.
func runUpdatePreflight(
	errW io.Writer,
	revision int,
	updated string,
	liveErr error,
	expectSet bool,
	expectRevision int,
) error {
	// --expect-revision runs first: an update that will be refused must not
	// print "proceeding" or "this update creates revision N+1" beforehand.
	if expectSet {
		if liveErr != nil {
			return fmt.Errorf(
				"--expect-revision %d: could not verify live revision: %w",
				expectRevision, liveErr,
			)
		}
		if err := preflight.CheckExpectedRevision(expectRevision, revision); err != nil {
			return err
		}
	}
	if liveErr != nil {
		preflight.PrintFailOpenNote(errW, "fetch live state", liveErr)
		return nil
	}
	preflight.PrintUpdateBanner(errW, "", revision, updated)
	return nil
}

// printConfigPreflight prints the content pin summary — and, with
// --show-diff, the full live-vs-incoming diff — for an update that replaces
// the whole agent config. rawIncoming is the config as it will be PATCHed.
// Best-effort like all pre-flight output: rendering problems never block.
func printConfigPreflight(
	errW io.Writer,
	liveConfig any,
	rawIncoming json.RawMessage,
	showDiff bool,
) {
	var incoming any
	if err := json.Unmarshal(rawIncoming, &incoming); err != nil {
		return
	}
	preflight.PrintPinChanges(errW, liveConfig, incoming)
	if showDiff {
		if err := output.PrintYAMLDiff(errW, "live", liveConfig, "incoming", incoming); err != nil {
			preflight.PrintFailOpenNote(errW, "render config diff", err)
		}
	}
}
