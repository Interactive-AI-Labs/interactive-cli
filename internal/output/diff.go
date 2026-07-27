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
	return PrintYAMLDiff(out, "revision "+nameA, mapA, "revision "+nameB, mapB)
}

// PrintYAMLDiff prints a unified diff between the YAML renderings of two
// values, labeled verbatim.
func PrintYAMLDiff(out io.Writer, labelA string, a any, labelB string, b any) error {
	yamlA, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", labelA, err)
	}
	yamlB, err := yaml.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", labelB, err)
	}

	if string(yamlA) == string(yamlB) {
		fmt.Fprintln(out, "No differences found.")
		return nil
	}

	var opts []textdiff.Option
	if IsTerminal(out) {
		opts = append(opts, textdiff.TerminalColors())
		fmt.Fprintf(out, "%s--- %s%s\n", colorRed, labelA, colorReset)
		fmt.Fprintf(out, "%s+++ %s%s\n", colorGreen, labelB, colorReset)
	} else {
		fmt.Fprintf(out, "--- %s\n", labelA)
		fmt.Fprintf(out, "+++ %s\n", labelB)
	}

	diff := textdiff.Unified(string(yamlA), string(yamlB), opts...)
	fmt.Fprint(out, diff)
	return nil
}
