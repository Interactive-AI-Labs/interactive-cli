package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/versioncheck"
	"github.com/spf13/cobra"
)

const (
	installPkg   = "github.com/Interactive-AI-Labs/interactive-cli/cmd/iai@latest"
	minGoVersion = "1.25.1"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update iai to the latest version",
	Long:  `Check for and install the latest version of the iai CLI using go install.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if _, err := exec.LookPath("go"); err != nil {
			return fmt.Errorf(
				"\"go\" is not installed or not in your PATH.\niai update requires Go to run \"go install\". Install Go from https://go.dev/dl/ or update iai manually",
			)
		}

		goVer := strings.TrimPrefix(runtime.Version(), "go")
		if !versioncheck.IsGoVersionSufficient(goVer, minGoVersion) {
			return fmt.Errorf(
				"iai requires Go %s or later, but you have Go %s.\nPlease upgrade Go from https://go.dev/dl/",
				minGoVersion,
				goVer,
			)
		}

		fmt.Fprintln(out, "Checking for updates...")

		latest, err := versioncheck.FetchLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to check latest version: %w", err)
		}

		if !versioncheck.IsNewer(version, latest) {
			fmt.Fprintf(out, "Already up to date (v%s).\n", version)
			return nil
		}

		fmt.Fprintf(out, "Updating iai: v%s → v%s\n", version, latest)

		install := exec.CommandContext(cmd.Context(), "go", "install", installPkg)
		install.Stdout = out
		install.Stderr = cmd.ErrOrStderr()
		if err := install.Run(); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Fprintf(out, "Successfully updated to v%s.\n", latest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
