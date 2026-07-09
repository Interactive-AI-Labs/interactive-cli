package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	replicasProject      string
	replicasOrganization string
	replicasListJSON     bool
	replicasListYAML     bool
	replicasDescribeJSON bool
	replicasDescribeYAML bool
)

var replicasCmd = &cobra.Command{
	Use:     "replicas",
	Short:   "Inspect service replicas",
	GroupID: groupInfra,
	Long:    `Manage pods backing services in a specific project.`,
}

var replicasListCmd = &cobra.Command{
	Use:     "list [service_name]",
	Aliases: []string{"ls"},
	Short:   "List replicas for a service",
	Long:    `List pods backing a service in a specific project.`,
	Example: `  iai replicas list my-service
  iai replicas list my-service -p my-project -o my-org
  iai replicas list my-service --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		serviceName := strings.TrimSpace(args[0])

		if serviceName == "" {
			return fmt.Errorf("service name is required")
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			replicasOrganization,
			replicasProject,
		)
		if err != nil {
			return err
		}

		replicas, err := deployClient.ListReplicas(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			serviceName,
		)
		if err != nil {
			return err
		}

		if replicasListJSON {
			return output.PrintStructuredJSON(out, replicas)
		}
		if replicasListYAML {
			return output.PrintStructuredYAML(out, replicas)
		}

		return output.PrintReplicaList(out, replicas)
	},
}

var replicasDescribeCmd = &cobra.Command{
	Use:     "describe <replica_name>",
	Aliases: []string{"desc"},
	Short:   "Describe a replica in detail",
	Long:    `Show detailed information about a specific replica including status, resources, healthcheck configuration, and events.`,
	Example: `  iai replicas describe my-service-abc123
  iai replicas describe my-service-abc123 -p my-project -o my-org
  iai replicas describe my-service-abc123 --yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		replicaName := strings.TrimSpace(args[0])

		if replicaName == "" {
			return fmt.Errorf("replica name is required")
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			replicasOrganization,
			replicasProject,
		)
		if err != nil {
			return err
		}

		status, err := deployClient.DescribeReplica(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			replicaName,
		)
		if err != nil {
			return err
		}

		if replicasDescribeJSON {
			return output.PrintStructuredJSON(out, status)
		}
		if replicasDescribeYAML {
			return output.PrintStructuredYAML(out, status)
		}

		return output.PrintReplicaDescribe(out, status)
	},
}

var (
	replicaLogsFollow     bool
	replicaLogsSince      string
	replicaLogsStartTime  string
	replicaLogsEndTime    string
	replicaLogsRaw        bool
	replicaLogsDecode     bool
	replicaLogsFields     []string
	replicaLogsAllFields  bool
	replicaLogsTimestamps bool
)

var replicasLogsCmd = &cobra.Command{
	Use:   "logs <replica_name>",
	Short: "Show logs for a specific replica",
	Long: `Show logs for a specific replica in a project.

Returns up to 1000 log entries in chronological order.

Structured (JSON) logs are automatically formatted: the level and message
fields are extracted and displayed as "LEVEL message". Use --fields or
--all-fields to include additional top-level fields after the message. Use
--raw for exact server JSON, or --decode to decode embedded JSON strings into
nested JSON values.`,
	Example: `  iai replicas logs my-service-abc123
  iai replicas logs my-service-abc123 --follow
  iai replicas logs my-service-abc123 --since 30m --fields logger,pid
  iai replicas logs my-service-abc123 --timestamps
  iai replicas logs my-service-abc123 --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z`,
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

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		timeout := 1 * time.Minute
		if replicaLogsFollow {
			timeout = 0
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			timeout,
			token,
			apiKey,
			cookies,
		)
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
			EndTime:   replicaLogsEndTime,
		}

		logsResp, err := deployClient.GetReplicaLogs(ctx, orgId, projectId, replicaName, opts)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		meta := output.LogsMeta{
			Start:     logsResp.Start,
			End:       logsResp.End,
			Truncated: logsResp.Truncated,
			Empty:     logsResp.Empty,
			Limit:     logsResp.Limit,
		}
		fmtOpts := output.LogFormatOptions{
			Raw:        replicaLogsRaw || replicaLogsDecode,
			Decode:     replicaLogsDecode,
			Fields:     replicaLogsFields,
			AllFields:  replicaLogsAllFields,
			Timestamps: replicaLogsTimestamps,
		}
		err = output.PrintLogStream(out, logsResp.Body, false, meta, fmtOpts)
		if replicaLogsFollow && ctx.Err() != nil {
			return nil
		}
		return err
	},
}

var replicaLogFieldsSince string

var replicaLogFieldsCmd = &cobra.Command{
	Use:   "log-fields <replica_name>",
	Short: "List available fields in structured logs",
	Long: `Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai replicas logs --fields' to include them in output.`,
	Example: `  iai replicas log-fields my-service-abc123
  iai replicas log-fields my-service-abc123 --since 1h`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		replicaName := strings.TrimSpace(args[0])

		since := replicaLogFieldsSince

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			replicasOrganization,
			replicasProject,
		)
		if err != nil {
			return err
		}

		opts := clients.LogsOptions{Since: since}
		logsResp, err := deployClient.GetReplicaLogs(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			replicaName,
			opts,
		)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		if logsResp.Empty {
			output.PrintNoLogsFound(cmd.ErrOrStderr(), logsResp.Start, logsResp.End)
			return nil
		}

		fields, err := output.DiscoverLogFields(logsResp.Body)
		if err != nil {
			return err
		}
		if err := output.PrintLogFields(out, fields); err != nil {
			return err
		}
		if logsResp.Truncated {
			output.PrintLogFieldDiscoveryTruncationWarning(cmd.ErrOrStderr(), logsResp.Limit)
		}
		return nil
	},
}

func init() {
	// Flags for "replicas list"
	replicasListCmd.Flags().
		StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasListCmd.Flags().
		StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")
	replicasListCmd.Flags().
		BoolVar(&replicasListJSON, "json", false, "Output raw API response as JSON")
	replicasListCmd.Flags().
		BoolVar(&replicasListYAML, "yaml", false, "Output raw API response as YAML")
	replicasListCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Flags for "replicas describe"
	replicasDescribeCmd.Flags().
		StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasDescribeCmd.Flags().
		StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")
	replicasDescribeCmd.Flags().
		BoolVar(&replicasDescribeJSON, "json", false, "Output raw API response as JSON")
	replicasDescribeCmd.Flags().
		BoolVar(&replicasDescribeYAML, "yaml", false, "Output raw API response as YAML")
	replicasDescribeCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Flags for "replicas logs"
	replicasLogsCmd.Flags().
		StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasLogsCmd.Flags().
		StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")
	replicasLogsCmd.Flags().
		BoolVarP(&replicaLogsFollow, "follow", "f", false, "Stream new log entries as they arrive; mutually exclusive with --end-time")
	replicasLogsCmd.Flags().
		StringVar(&replicaLogsSince, "since", "", "Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time")
	replicasLogsCmd.Flags().
		StringVar(&replicaLogsStartTime, "start-time", "", "Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window")
	replicasLogsCmd.Flags().
		StringVar(&replicaLogsEndTime, "end-time", "", "Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow")
	replicasLogsCmd.Flags().
		BoolVar(&replicaLogsRaw, "raw", false, "Output exact server JSON lines without formatting")
	replicasLogsCmd.Flags().
		BoolVar(&replicaLogsDecode, "decode", false, "Decode embedded JSON strings into nested JSON values; outputs raw JSON")
	replicasLogsCmd.Flags().
		StringSliceVar(&replicaLogsFields, "fields", nil, "Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON")
	replicasLogsCmd.Flags().
		BoolVar(&replicaLogsAllFields, "all-fields", false, "Show all extra top-level fields from structured (JSON) logs after the message")
	replicasLogsCmd.Flags().
		BoolVar(&replicaLogsTimestamps, "timestamps", false, "Include platform log timestamps")
	replicasLogsCmd.MarkFlagsMutuallyExclusive("raw", "fields")
	replicasLogsCmd.MarkFlagsMutuallyExclusive("raw", "all-fields")
	replicasLogsCmd.MarkFlagsMutuallyExclusive("decode", "fields")
	replicasLogsCmd.MarkFlagsMutuallyExclusive("decode", "all-fields")
	replicasLogsCmd.MarkFlagsMutuallyExclusive("fields", "all-fields")

	// Flags for "replicas log-fields"
	replicaLogFieldsCmd.Flags().
		StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicaLogFieldsCmd.Flags().
		StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")
	replicaLogFieldsCmd.Flags().
		StringVar(&replicaLogFieldsSince, "since", "1h", "Relative duration to scan (e.g. 5m, 1h)")

	replicasCmd.AddCommand(replicasListCmd)
	replicasCmd.AddCommand(replicasDescribeCmd)
	replicasCmd.AddCommand(replicasLogsCmd)
	replicasCmd.AddCommand(replicaLogFieldsCmd)
	rootCmd.AddCommand(replicasCmd)
}
