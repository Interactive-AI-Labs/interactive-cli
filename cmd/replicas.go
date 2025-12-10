package cmd

import (
	"fmt"
	"strings"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var (
	replicasProject      string
	replicasOrganization string
)

var replicasCmd = &cobra.Command{
	Use:   "replicas",
	Short: "Manage service replicas",
	Long:  `Manage pods backing services in a specific project.`,
}

var replicasListCmd = &cobra.Command{
	Use:     "list [service_name]",
	Aliases: []string{"ls"},
	Short:   "List replicas for a service",
	Long: `List pods backing a service in a specific project.

The project is selected with --project.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		serviceName := strings.TrimSpace(args[0])

		if strings.TrimSpace(replicasProject) == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required")
		}

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
		if strings.TrimSpace(replicasOrganization) == "" {
			if strings.TrimSpace(selectedOrg) == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			replicasOrganization = selectedOrg
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), replicasOrganization, replicasProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", replicasProject, err)
		}

		replicas, err := deployClient.ListReplicas(cmd.Context(), orgId, projectId, serviceName)
		if err != nil {
			return err
		}

		headers := []string{"NAME", "STATUS", "CPU", "MEMORY", "STARTED"}
		rows := make([][]string, len(replicas))
		for i, r := range replicas {
			readinessLabel := "Not Ready"
			if r.Ready {
				readinessLabel = "Ready"
			}

			combinedStatus := strings.TrimSpace(r.Status)
			if combinedStatus == "" {
				combinedStatus = strings.TrimSpace(r.Phase)
			}
			if combinedStatus == "" {
				combinedStatus = "Unknown"
			}

			combinedStatus = fmt.Sprintf("%s [%s]", combinedStatus, readinessLabel)

			rows[i] = []string{
				r.Name,
				combinedStatus,
				r.CPU,
				r.Memory,
				r.StartTime,
			}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

func init() {
	// Flags for "replicas list"
	replicasListCmd.Flags().StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasListCmd.Flags().StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")

	replicasCmd.AddCommand(replicasListCmd)
	rootCmd.AddCommand(replicasCmd)
}
