package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func validateTableOnlyColumns(cmd *cobra.Command, asJSON, asYAML bool) error {
	if !cmd.Flags().Changed("columns") || (!asJSON && !asYAML) {
		return nil
	}

	return fmt.Errorf("--columns cannot be used with --json or --yaml")
}
