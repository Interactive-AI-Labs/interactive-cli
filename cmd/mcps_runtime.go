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
	Short: "Show logs for an MCP",
	Long: `Fetch or follow logs produced by an internal MCP's replicas through the deployment operator.
External provider logs are not collected. Requests query retained logs by MCP name.

Returns up to 1000 entries by default; --limit accepts up to 5000. Structured
logs are formatted as "LEVEL message". Use --fields or --all-fields to include
extra fields, --raw for server JSON lines, or --decode for decoded JSON.

Use --summary for JSON severity counts over the entire time window, without
the log-entry limit. Unlabeled lines contribute to total only.`,
	Example: `  iai mcps logs my-tool --follow
  iai mcps logs my-tool --since 30m --level error --message 'timeout|failed'
  iai mcps logs my-tool --fields logger,pid --timestamps
  iai mcps logs my-tool --summary --since 3h
  iai mcps logs my-tool --from-timestamp 2026-01-01T00:00:00Z --to-timestamp 2026-01-01T01:00:00Z`,
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
	Use:     "log-fields <mcp_name>",
	Short:   "List available fields in structured MCP logs",
	Long:    "Scan recent logs for extra top-level JSON fields to use with 'iai mcps logs --fields'.",
	Example: "  iai mcps log-fields my-tool --since 1h",
	Args:    cobra.ExactArgs(1),
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

func init() {
	f := mcpLogsCmd.Flags()
	f.BoolVarP(&mcpLogsOptions.Follow, "follow", "f", false, "Stream new log entries")
	f.StringVar(
		&mcpLogsOptions.Since,
		"since",
		"",
		"Relative lookback (e.g. 30m, 1h, 3d); default 1h; maximum 72h",
	)
	f.StringVar(
		&mcpLogsOptions.StartTime,
		"from-timestamp",
		"",
		"RFC3339 start timestamp; mutually exclusive with --since",
	)
	f.StringVar(
		&mcpLogsOptions.EndTime,
		"to-timestamp",
		"",
		"RFC3339 end timestamp; requires --from-timestamp; mutually exclusive with --since and --follow",
	)
	f.IntVar(&mcpLogsOptions.Limit, "limit", 0, "Maximum log entries (1-5000); defaults to 1000")
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
	f.BoolVar(&mcpLogsFormat.Decode, "decode", false, "Decode embedded JSON strings; output JSON")
	f.StringSliceVar(
		&mcpLogsFormat.Fields,
		"fields",
		nil,
		"Extra structured log fields to display (e.g. logger,pid)",
	)
	f.BoolVar(
		&mcpLogsFormat.AllFields,
		"all-fields",
		false,
		"Display all extra structured log fields",
	)
	f.BoolVar(&mcpLogsFormat.Timestamps, "timestamps", false, "Include platform log timestamps")
	for _, flag := range []string{"follow", "limit", "level", "raw", "decode", "fields", "all-fields", "timestamps"} {
		mcpLogsCmd.MarkFlagsMutuallyExclusive("summary", flag)
	}
	mcpLogsCmd.MarkFlagsMutuallyExclusive("since", "from-timestamp")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("since", "to-timestamp")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("follow", "to-timestamp")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("raw", "fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("raw", "all-fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("decode", "fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("decode", "all-fields")
	mcpLogsCmd.MarkFlagsMutuallyExclusive("fields", "all-fields")
	mcpLogFieldsCmd.Flags().
		StringVar(&mcpLogFieldsSince, "since", "1h", "Relative duration to scan (e.g. 5m, 1h)")
	mcpsCmd.AddCommand(mcpLogsCmd, mcpLogFieldsCmd)
}
