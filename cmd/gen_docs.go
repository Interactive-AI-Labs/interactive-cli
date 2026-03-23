package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const defaultDocsOutDir = "./docs"

//go:embed templates/install.md
var installInstructions string

//go:embed templates/sync_config_example.md
var syncConfigExample string

//go:embed templates/schema_routines.md
var schemaRoutines string

//go:embed templates/schema_policies.md
var schemaPolicies string

//go:embed templates/schema_variables.md
var schemaVariables string

//go:embed templates/schema_glossaries.md
var schemaGlossaries string

//go:embed templates/schema_macros.md
var schemaMacros string

// promptSchemaInserts maps doc filenames to their markdown schema templates.
// These replace the plain-text schema sections generated from the cobra Long
// descriptions with properly fenced code blocks for the docs.
var promptSchemaInserts = map[string]string{
	"iai_routines_create.md":   schemaRoutines,
	"iai_routines_update.md":   schemaRoutines,
	"iai_policies_create.md":   schemaPolicies,
	"iai_policies_update.md":   schemaPolicies,
	"iai_variables_create.md":  schemaVariables,
	"iai_variables_update.md":  schemaVariables,
	"iai_glossaries_create.md": schemaGlossaries,
	"iai_glossaries_update.md": schemaGlossaries,
	"iai_macros_create.md":     schemaMacros,
	"iai_macros_update.md":     schemaMacros,
}

var genDocsCmd = &cobra.Command{
	Use:    "gen-docs",
	Short:  "Generate Markdown docs for all commands",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		outDir, err := cmd.Flags().GetString("out-dir")
		if err != nil {
			return err
		}
		if outDir == "" {
			outDir = defaultDocsOutDir
		}

		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}

		rootCmd.DisableAutoGenTag = true

		err = doc.GenMarkdownTreeCustom(
			rootCmd,
			outDir,
			func(string) string { return "" },
			func(name string) string { return name },
		)
		if err != nil {
			return err
		}

		rootDocPath := filepath.Join(outDir, "iai.md")
		content, err := os.ReadFile(rootDocPath)
		if err != nil {
			return err
		}

		// Find the Synopsis section and insert after it
		docContent := string(content)
		synopsisIdx := strings.Index(docContent, "### Synopsis")
		if synopsisIdx != -1 {
			nextSectionIdx := strings.Index(docContent[synopsisIdx+1:], "### ")
			if nextSectionIdx != -1 {
				insertPos := synopsisIdx + 1 + nextSectionIdx
				docContent = docContent[:insertPos] + installInstructions + "\n" + docContent[insertPos:]
			}
		}

		// Insert config example into the stack sync doc
		syncDocPath := filepath.Join(outDir, "iai_stack_sync.md")
		syncContent, err := os.ReadFile(syncDocPath)
		if err != nil {
			return err
		}
		syncDoc := string(syncContent)
		optionsIdx := strings.Index(syncDoc, "### Options")
		if optionsIdx != -1 {
			syncDoc = syncDoc[:optionsIdx] + syncConfigExample + "\n" + syncDoc[optionsIdx:]
			if err := os.WriteFile(syncDocPath, []byte(syncDoc), 0o644); err != nil {
				return err
			}
		}

		if err := os.WriteFile(rootDocPath, []byte(docContent), 0o644); err != nil {
			return err
		}

		// Replace plain-text schema sections in prompt type docs with
		// properly formatted markdown (fenced code blocks).
		for filename, schemaDoc := range promptSchemaInserts {
			if err := injectSchemaDoc(outDir, filename, schemaDoc); err != nil {
				return fmt.Errorf("failed to inject schema into %s: %w", filename, err)
			}
		}

		return nil
	},
}

// injectSchemaDoc replaces the plain-text schema/example section in a
// generated prompt type doc with a markdown-formatted version.
func injectSchemaDoc(outDir, filename, schemaDoc string) error {
	docPath := filepath.Join(outDir, filename)
	content, err := os.ReadFile(docPath)
	if err != nil {
		return err
	}

	text := string(content)

	// Find the start of the schema section. The marker text may appear
	// mid-line (e.g. "...follow the routine schema. Run 'iai ..."), so we
	// search for the containing line and use its start.
	markers := []string{"Run 'iai", "No schema"}
	schemaStart := -1
	for _, m := range markers {
		idx := strings.Index(text, m)
		if idx == -1 {
			continue
		}
		// Walk back to the start of the line.
		lineStart := strings.LastIndex(text[:idx], "\n")
		if lineStart == -1 {
			lineStart = 0
		}
		schemaStart = lineStart
		break
	}
	if schemaStart == -1 {
		return fmt.Errorf("schema section marker not found in %s", filename)
	}

	// Find "### Options" which follows the schema section
	optionsMarker := "\n### Options"
	optionsIdx := strings.Index(text, optionsMarker)
	if optionsIdx == -1 {
		return fmt.Errorf("options section marker not found in %s", filename)
	}

	// Replace the plain-text section with the markdown template
	text = text[:schemaStart] + "\n\n" + schemaDoc + optionsMarker + text[optionsIdx+len(optionsMarker):]

	return os.WriteFile(docPath, []byte(text), 0o644)
}

func init() {
	genDocsCmd.Flags().
		String("out-dir", defaultDocsOutDir, "Output directory for generated Markdown docs")
	rootCmd.AddCommand(genDocsCmd)
}
