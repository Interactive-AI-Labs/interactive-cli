package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/preflight"
)

func runUpdatePreflight(
	errW io.Writer,
	revision int,
	updated string,
	liveErr error,
	expectSet bool,
	expectRevision int,
) error {
	if expectSet {
		if liveErr != nil {
			return fmt.Errorf(
				"--expect-revision %d: could not verify live revision: %w",
				expectRevision, liveErr,
			)
		}
		if revision != expectRevision {
			return fmt.Errorf(
				"live revision is %d, expected %d — not applying (--expect-revision)",
				revision, expectRevision,
			)
		}
	}
	if liveErr != nil {
		preflight.PrintFailOpenNote(errW, "fetch live state", liveErr)
		return nil
	}
	preflight.PrintUpdateBanner(errW, "", revision, updated)
	return nil
}

func printDroppedEnvSecretWarnings(
	errW io.Writer,
	envChanged, secretChanged bool,
	liveEnv []clients.EnvVar,
	liveRefs []clients.SecretRef,
	envArgs, secretArgs []string,
) bool {
	dropped := false
	if envChanged {
		live := make([]string, 0, len(liveEnv))
		for _, e := range liveEnv {
			live = append(live, e.Name)
		}
		incoming := make([]string, 0, len(envArgs))
		for _, e := range envArgs {
			incoming = append(incoming, strings.TrimSpace(strings.SplitN(e, "=", 2)[0]))
		}
		if preflight.PrintDroppedListEntries(errW, "env vars", "--env", live, incoming) {
			dropped = true
		}
	}
	if secretChanged {
		live := make([]string, 0, len(liveRefs))
		for _, r := range liveRefs {
			live = append(live, r.SecretName)
		}
		incoming := make([]string, 0, len(secretArgs))
		for _, s := range secretArgs {
			incoming = append(incoming, strings.TrimSpace(s))
		}
		if preflight.PrintDroppedListEntries(errW, "secret refs", "--secret", live, incoming) {
			dropped = true
		}
	}
	return dropped
}

func checkUpdateGates(force, pinRollback, droppedEntries bool) error {
	var reasons []string
	if pinRollback {
		reasons = append(reasons, "downgrades or removes live content pins")
	}
	if droppedEntries {
		reasons = append(reasons, "drops live env vars or secret refs")
	}
	if len(reasons) == 0 || force {
		return nil
	}
	return fmt.Errorf(
		"refusing to apply: this update %s (details above) — pass --force if this is intended",
		strings.Join(reasons, " and "),
	)
}

func printConfigPreflight(
	errW io.Writer,
	liveConfig any,
	rawIncoming json.RawMessage,
	showDiff bool,
) bool {
	var incoming any
	if err := json.Unmarshal(rawIncoming, &incoming); err != nil {
		return false
	}
	pinRollback := preflight.PrintPinChanges(errW, liveConfig, incoming)
	if showDiff {
		if err := output.PrintYAMLDiff(errW, "live", liveConfig, "incoming", incoming); err != nil {
			preflight.PrintFailOpenNote(errW, "render config diff", err)
		}
	}
	return pinRollback
}
