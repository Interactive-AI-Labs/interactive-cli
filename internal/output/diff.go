package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
	"znkr.io/diff/textdiff"
)

var revisionMetaKeys = []string{"revision", "updated", "status"}

func stripRevisionMeta(v any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	for _, k := range revisionMetaKeys {
		delete(m, k)
	}
	return m, nil
}

// PrintRevisionDiff prints a colored unified diff between two revision snapshots.
// RevisionMeta fields (revision, updated, status) are excluded.
func PrintRevisionDiff(out io.Writer, nameA string, a any, nameB string, b any) error {
	mapA, err := stripRevisionMeta(a)
	if err != nil {
		return fmt.Errorf("failed to process revision %s: %w", nameA, err)
	}
	mapB, err := stripRevisionMeta(b)
	if err != nil {
		return fmt.Errorf("failed to process revision %s: %w", nameB, err)
	}

	yamlA, err := yaml.Marshal(mapA)
	if err != nil {
		return fmt.Errorf("failed to marshal revision %s: %w", nameA, err)
	}
	yamlB, err := yaml.Marshal(mapB)
	if err != nil {
		return fmt.Errorf("failed to marshal revision %s: %w", nameB, err)
	}

	if string(yamlA) == string(yamlB) {
		fmt.Fprintln(out, "No differences found.")
		return nil
	}

	var opts []textdiff.Option
	if IsTerminal(out) {
		opts = append(opts, textdiff.TerminalColors())
		fmt.Fprintf(out, "%s--- revision %s%s\n", colorRed, nameA, colorReset)
		fmt.Fprintf(out, "%s+++ revision %s%s\n", colorGreen, nameB, colorReset)
	} else {
		fmt.Fprintf(out, "--- revision %s\n", nameA)
		fmt.Fprintf(out, "+++ revision %s\n", nameB)
	}

	diff := textdiff.Unified(string(yamlA), string(yamlB), opts...)
	fmt.Fprint(out, diff)
	return nil
}
