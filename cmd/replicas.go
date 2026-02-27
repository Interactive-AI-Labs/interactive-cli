package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

		return output.PrintReplicaList(out, replicas)
	},
}

var replicasDescribeCmd = &cobra.Command{
	Use:     "describe <replica_name>",
	Aliases: []string{"desc"},
	Short:   "Describe a replica in detail",
	Long: `Show detailed information about a specific replica including status, resources, healthcheck configuration, and events.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		replicaName := strings.TrimSpace(args[0])

		if replicaName == "" {
			return fmt.Errorf("replica name is required")
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

		status, err := deployClient.DescribeReplica(cmd.Context(), orgId, projectId, replicaName)
		if err != nil {
			return err
		}

		return output.PrintReplicaDescribe(out, status)
	},
}

var (
	replicaLogsFollow    bool
	replicaLogsSince     string
	replicaLogsStartTime string
)

var replicasLogsCmd = &cobra.Command{
	Use:   "logs <replica_name>",
	Short: "Show logs for a specific replica",
	Long: `Show logs for a specific replica in a project.

Returns up to 5000 log entries in chronological order. Default lookback is 1h.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		replicaName := strings.TrimSpace(args[0])
		if replicaName == "" {
			return fmt.Errorf("replica name is required")
		}

		ctx := cmd.Context()
		if replicaLogsFollow {
			var stop func()
			ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
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

		timeout := 1 * time.Minute
		if replicaLogsFollow {
			timeout = 0
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, timeout, apiKey, cookies)
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

		opts := clients.LogsOptions{
			Follow:    replicaLogsFollow,
			Since:     replicaLogsSince,
			StartTime: replicaLogsStartTime,
		}

		logsResp, err := deployClient.GetReplicaLogs(ctx, orgId, projectId, replicaName, opts)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		meta := output.LogsMeta{Since: logsResp.Since, Truncated: logsResp.Truncated}
		err = output.PrintLogStream(out, logsResp.Body, false, meta)
		if replicaLogsFollow && ctx.Err() != nil {
			return nil
		}
		return err
	},
}

func init() {
	// Flags for "replicas list"
	replicasListCmd.Flags().StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasListCmd.Flags().StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "replicas describe"
	replicasDescribeCmd.Flags().StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasDescribeCmd.Flags().StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "replicas logs"
	replicasLogsCmd.Flags().StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasLogsCmd.Flags().StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")
	replicasLogsCmd.Flags().BoolVarP(&replicaLogsFollow, "follow", "f", false, "Follow log output")
	replicasLogsCmd.Flags().StringVar(&replicaLogsSince, "since", "", "Relative duration to look back (e.g. 5m, 1h, 3d); default 1h, max 3d")
	replicasLogsCmd.Flags().StringVar(&replicaLogsStartTime, "start-time", "", "Absolute RFC3339 timestamp to start from (e.g. 2026-02-24T10:00:00Z); max 3d ago, mutually exclusive with --since")

	replicasCmd.AddCommand(replicasListCmd)
	replicasCmd.AddCommand(replicasDescribeCmd)
	replicasCmd.AddCommand(replicasLogsCmd)
	rootCmd.AddCommand(replicasCmd)
}
