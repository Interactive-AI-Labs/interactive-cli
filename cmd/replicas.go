package cmd

import (
	"fmt"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
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

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		serviceName := strings.TrimSpace(args[0])

		if serviceName == "" {
			return fmt.Errorf("service name is required")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, replicasOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, replicasProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
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

		if err := output.PrintTable(out, headers, rows); err != nil {
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
