package cmd

import (
	"fmt"
	"io"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var (
	logsProject      string
	logsOrganization string
	logsFollow       bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [replica_name]",
	Short: "Show logs for a specific replica",
	Long: `Show logs for a specific replica (pod) in a project.

The project is selected with --project. If --organization is not provided,
the currently selected organization (via 'interactiveai organizations select')
is used.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if logsProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}

		replicaName := args[0]

		// Load session cookies.
		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := internal.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := internal.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if logsOrganization == "" {
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			logsOrganization = selectedOrg
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), logsOrganization, logsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", logsProject, err)
		}

		logReader, err := deployClient.GetLogs(cmd.Context(), orgId, projectId, replicaName, logsFollow)
		if err != nil {
			return err
		}
		defer logReader.Close()

		_, err = io.Copy(out, logReader)
		return err
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsProject, "project", "p", "", "Project name that owns the service")
	logsCmd.Flags().StringVarP(&logsOrganization, "organization", "o", "", "Organization name that owns the project")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")

	rootCmd.AddCommand(logsCmd)
}
