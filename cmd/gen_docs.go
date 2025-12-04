package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const defaultDocsOutDir = "./docs"

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

		return doc.GenMarkdownTree(rootCmd, outDir)
	},
}

func init() {
	genDocsCmd.Flags().String("out-dir", defaultDocsOutDir, "Output directory for generated Markdown docs")
	rootCmd.AddCommand(genDocsCmd)
}
