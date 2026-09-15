package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	mcpLogsOptions    deployment.LogsOptions
	mcpLogsFormat     output.LogFormatOptions
	mcpLogsSummary    bool
	mcpLogFieldsSince string
)

var mcpLogsCmd = &cobra.Command{
	Use:   "logs <mcp_name>",
	Short: "Show logs for an mcp",
	Long: `Show logs for all replicas of an mcp in a project.

Returns up to 1000 log entries in chronological order by default; use
--limit to request up to 5000.

Structured (JSON) logs are automatically formatted: the level and message
fields are extracted and displayed as "LEVEL message". Use --fields or
--all-fields to include additional top-level fields after the message. Use
--raw for exact server JSON, or --decode to decode embedded JSON strings into
nested JSON values.

Use --summary for JSON severity counts over the entire time window, without
the log-entry limit. Unlabeled lines contribute to total only.`,
	Example: `  iai mcps logs my-tool
  iai mcps logs my-tool --follow
  iai mcps logs my-tool --since 3h
  iai mcps logs my-tool --timestamps
  iai mcps logs my-tool --fields logger,pid
  iai mcps logs my-tool --since 30m --level error --message 'timeout|failed'
  iai mcps logs my-tool --summary --since 3h
  iai mcps logs my-tool --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("mcp name is required")
		}

		ctx := cmd.Context()
		timeout := time.Minute
		if mcpLogsOptions.Follow {
			var stop func()
			ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			timeout = 0
		}

		pCtx, _, client, err := resolveProject(
			ctx,
			mcpOrganization,
			mcpProject,
			resolveOpts{deployTimeout: timeout},
		)
		if err != nil {
			return err
		}

		if mcpLogsSummary {
			raw, err := client.GetMcpLogSummary(
				ctx,
				pCtx.orgId,
				pCtx.projectId,
				name,
				mcpLogsOptions,
			)
			if err != nil {
				return err
			}
			return output.PrintRawJSON(cmd.OutOrStdout(), raw)
		}

		logs, err := client.GetMcpLogs(ctx, pCtx.orgId, pCtx.projectId, name, mcpLogsOptions)
		if err != nil {
			if mcpLogsOptions.Follow && ctx.Err() != nil {
				return nil
			}
			return err
		}
		defer logs.Body.Close()

		format := mcpLogsFormat
		format.Raw = format.Raw || format.Decode
		err = output.PrintLogStream(cmd.OutOrStdout(), logs.Body, true, output.LogsMeta{
			Start:     logs.Start,
			End:       logs.End,
			Truncated: logs.Truncated,
			Empty:     logs.Empty,
			Limit:     logs.Limit,
		}, format)
		if mcpLogsOptions.Follow && ctx.Err() != nil {
			return nil
		}
		return err
	},
}

var mcpLogFieldsCmd = &cobra.Command{
	Use:   "log-fields <mcp_name>",
	Short: "List available fields in structured logs",
	Long: `Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai mcps logs --fields' to include them in output.`,
	Example: `  iai mcps log-fields my-tool
  iai mcps log-fields my-tool --since 1h`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pCtx, _, client, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		logs, err := client.GetMcpLogs(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			strings.TrimSpace(args[0]),
			deployment.LogsOptions{Since: mcpLogFieldsSince},
		)
		if err != nil {
			return err
		}
		defer logs.Body.Close()

		if logs.Empty {
			output.PrintNoLogsFound(cmd.ErrOrStderr(), logs.Start, logs.End)
			return nil
		}

		fields, err := output.DiscoverLogFields(logs.Body)
		if err != nil {
			return err
		}
		if err := output.PrintLogFields(cmd.OutOrStdout(), fields); err != nil {
			return err
		}

		if logs.Truncated {
			output.PrintLogFieldDiscoveryTruncationWarning(cmd.ErrOrStderr(), logs.Limit)
		}
		return nil
	},
}

var mcpRestartCmd = &cobra.Command{
	Use:   "restart <mcp_name>",
	Short: "Restart an mcp in a project",
	Long:  `Restart an mcp in a specific project using the deployment service.`,
	Example: `  iai mcps restart my-tool
  iai mcps restart my-tool --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: runMcpAction,
}

var mcpActivateCmd = &cobra.Command{
	Use:   "activate <mcp_name>",
	Short: "Activate a deactivated mcp in a project",
	Long:  `Activate a deactivated mcp, restoring it to its previous configuration.`,
	Example: `  iai mcps activate my-tool
  iai mcps activate my-tool --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: runMcpAction,
}

var mcpDeactivateCmd = &cobra.Command{
	Use:   "deactivate <mcp_name>",
	Short: "Deactivate an mcp in a project",
	Long: `Deactivate an mcp, stopping all running instances. The current configuration
is preserved and will be restored when the mcp is activated again.`,
	Example: `  iai mcps deactivate my-tool
  iai mcps deactivate my-tool --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: runMcpAction,
}

func runMcpAction(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	name := strings.TrimSpace(args[0])
	if name == "" {
		return fmt.Errorf("mcp name is required")
	}

	pCtx, _, client, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Submitting mcp %s request...\n", cmd.Name())

	message, err := client.McpAction(cmd.Context(), pCtx.orgId, pCtx.projectId, name, cmd.Name())
	if err != nil {
		return err
	}

	if message != "" {
		fmt.Fprintln(out, message)
	}

	return nil
}

func init() {
	mcpsCmd.AddCommand(mcpRestartCmd, mcpActivateCmd, mcpDeactivateCmd)

	f := mcpLogsCmd.Flags()
	f.BoolVarP(
		&mcpLogsOptions.Follow,
		"follow",
		"f",
		false,
		"Stream new log entries as they arrive; mutually exclusive with --end-time",
	)
	f.StringVar(
		&mcpLogsOptions.Since,
		"since",
		"",
		"Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time",
	)
	f.StringVar(
		&mcpLogsOptions.StartTime,
		"start-time",
		"",
		"Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window",
	)
	f.StringVar(
		&mcpLogsOptions.EndTime,
		"end-time",
		"",
		"Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow",
	)
	f.IntVar(
		&mcpLogsOptions.Limit,
		"limit",
		0,
		"Maximum number of log entries to return (1-5000); defaults to 1000",
	)
	f.StringVar(
		&mcpLogsOptions.Message,
		"message",
		"",
		"Case-insensitive RE2 regular expression matched against the log line",
	)
	f.StringVar(
		&mcpLogsOptions.Level,
		"level",
		"",
		"JSON log severity: debug, info, warn, or error",
	)
	f.BoolVar(
		&mcpLogsSummary,
		"summary",
		false,
		"Output JSON severity counts instead of log entries",
	)
	f.BoolVar(&mcpLogsFormat.Raw, "raw", false, "Output exact server JSON lines without formatting")
	f.BoolVar(
		&mcpLogsFormat.Decode,
		"decode",
		false,
		"Decode embedded JSON strings into nested JSON values; outputs raw JSON",
	)
	f.StringSliceVar(
		&mcpLogsFormat.Fields,
		"fields",
		nil,
		"Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON",
	)
	f.BoolVar(
		&mcpLogsFormat.AllFields,
		"all-fields",
		false,
		"Show all extra top-level fields from structured (JSON) logs after the message",
	)
	f.BoolVar(&mcpLogsFormat.Timestamps, "timestamps", false, "Include platform log timestamps")

	for _, flag := range []string{"follow", "limit", "level", "raw", "decode", "fields", "all-fields", "timestamps"} {
		mcpLogsCmd.MarkFlagsMutuallyExclusive("summary", flag)
	}
	mcpLogsCmd.MarkFlagsMutuallyExclusive("raw", "fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("raw", "all-fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("decode", "fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("decode", "all-fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("fields", "all-fields")

	mcpLogFieldsCmd.Flags().
		StringVar(&mcpLogFieldsSince, "since", "1h", "Relative duration to scan (e.g. 5m, 1h)")

	mcpsCmd.AddCommand(mcpLogsCmd, mcpLogFieldsCmd)
}
