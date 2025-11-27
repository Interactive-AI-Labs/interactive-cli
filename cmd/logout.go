package cmd

import (
	"fmt"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of InteractiveAI",
	Long:  `Log out of InteractiveAI by clearing the local session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if err := internal.DeleteSessionCookies(cfgDirName, sessionFileName); err != nil {
			return fmt.Errorf("failed to logout: %w", err)
		}

		fmt.Fprintln(out, "Logged out successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
