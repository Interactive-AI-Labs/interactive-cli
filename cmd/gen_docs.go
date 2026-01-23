package cmd

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const defaultDocsOutDir = "./docs"

//go:embed templates/install.md
var installInstructions string

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

		// Disable auto-generated timestamp in docs
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

		return os.WriteFile(rootDocPath, []byte(docContent), 0o644)
	},
}

func init() {
	genDocsCmd.Flags().String("out-dir", defaultDocsOutDir, "Output directory for generated Markdown docs")
	rootCmd.AddCommand(genDocsCmd)
}
