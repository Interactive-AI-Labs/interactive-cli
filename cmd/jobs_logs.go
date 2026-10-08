package cmd

import (
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
	jobLogsFollow                                                  bool
	jobLogsSince, jobLogsStartTime, jobLogsEndTime                 string
	jobLogsRaw, jobLogsDecode, jobLogsAllFields, jobLogsTimestamps bool
	jobLogsFields                                                  []string
	jobLogsLimit                                                   int
	jobLogFieldsSince                                              string
)

var jobLogsCmd = &cobra.Command{
	Use:   "logs <job_name>",
	Short: "Show logs for a job",
	Long: `Show logs across a job's runs with 8-character run ID prefixes.
Log retention is independent of run history.`,
	Example: `  iai jobs logs my-job
  iai jobs logs my-job --follow
  iai jobs logs my-job --since 3h
  iai jobs logs my-job --timestamps
  iai jobs logs my-job --fields logger,pid`,
	Args: cobra.ExactArgs(1),
	RunE: runJobLogs,
}

var jobRunLogsCmd = &cobra.Command{
	Use:   "logs <run_id>",
	Short: "Show logs for one job run",
	Long:  `Show one run's logs, including after run-history cleanup.`,
	Example: `  iai jobs runs logs <run_id>
  iai jobs runs logs <run_id> --follow
  iai jobs runs logs <run_id> --timestamps`,
	Args: cobra.ExactArgs(1),
	RunE: runJobLogs,
}

func runJobLogs(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	nameOrID := strings.TrimSpace(args[0])
	singleRun := cmd.Parent() == jobRunsCmd

	ctx := cmd.Context()
	if jobLogsFollow {
		var stop func()
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}

	timeout := time.Minute
	if jobLogsFollow {
		timeout = 0
	}
	pCtx, _, client, err := resolveProject(
		ctx,
		jobOrganization,
		jobProject,
		resolveOpts{deployTimeout: timeout},
	)
	if err != nil {
		return err
	}

	opts := deployment.LogsOptions{
		Follow:    jobLogsFollow,
		Since:     jobLogsSince,
		StartTime: jobLogsStartTime,
		EndTime:   jobLogsEndTime,
		Limit:     jobLogsLimit,
	}
	getLogs := client.GetJobLogs
	if singleRun {
		getLogs = client.GetJobRunLogs
	}
	logs, err := getLogs(ctx, pCtx.orgId, pCtx.projectId, nameOrID, opts)
	if err != nil {
		return err
	}
	defer logs.Body.Close()

	meta := output.LogsMeta{
		Start:     logs.Start,
		End:       logs.End,
		Truncated: logs.Truncated,
		Empty:     logs.Empty,
		Limit:     logs.Limit,
	}
	format := output.LogFormatOptions{
		Raw:        jobLogsRaw || jobLogsDecode,
		Decode:     jobLogsDecode,
		Fields:     jobLogsFields,
		AllFields:  jobLogsAllFields,
		Timestamps: jobLogsTimestamps,
	}
	err = output.PrintLogStream(out, logs.Body, !singleRun, meta, format)
	if jobLogsFollow && ctx.Err() != nil {
		return nil
	}
	return err
}

var jobLogFieldsCmd = &cobra.Command{
	Use:   "log-fields <job_name>",
	Short: "List available fields in structured logs",
	Long: `Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai jobs logs --fields' to include them in output.`,
	Example: `  iai jobs log-fields my-job
  iai jobs log-fields my-job --since 1h`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		logs, err := client.GetJobLogs(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			name,
			deployment.LogsOptions{Since: jobLogFieldsSince},
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
	for _, cmd := range []*cobra.Command{jobLogsCmd, jobRunLogsCmd} {
		flags := cmd.Flags()
		flags.BoolVarP(
			&jobLogsFollow,
			"follow",
			"f",
			false,
			"Stream new entries for up to 10 minutes; reconnect to continue; cannot combine with --end-time",
		)
		flags.StringVar(
			&jobLogsSince,
			"since",
			"",
			logsSinceUsage,
		)
		flags.StringVar(
			&jobLogsStartTime,
			"start-time",
			"",
			logsStartTimeUsage,
		)
		flags.StringVar(
			&jobLogsEndTime,
			"end-time",
			"",
			"Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow",
		)
		flags.BoolVar(
			&jobLogsRaw,
			"raw",
			false,
			"Output exact server JSON lines without formatting",
		)
		flags.BoolVar(
			&jobLogsDecode, "decode", false,
			"Decode embedded JSON strings into nested JSON values; outputs raw JSON",
		)
		flags.StringSliceVar(
			&jobLogsFields,
			"fields",
			nil,
			"Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON",
		)
		flags.BoolVar(
			&jobLogsAllFields, "all-fields", false,
			"Show all extra top-level fields from structured (JSON) logs after the message",
		)
		flags.BoolVar(&jobLogsTimestamps, "timestamps", false, "Include platform log timestamps")
		flags.IntVar(
			&jobLogsLimit, "limit", 0,
			logsLimitUsage+"; with --follow, limits only the initial batch",
		)
		cmd.MarkFlagsMutuallyExclusive("raw", "fields")
		cmd.MarkFlagsMutuallyExclusive("raw", "all-fields")
		cmd.MarkFlagsMutuallyExclusive("decode", "fields")
		cmd.MarkFlagsMutuallyExclusive("decode", "all-fields")
		cmd.MarkFlagsMutuallyExclusive("fields", "all-fields")
	}
	jobLogFieldsCmd.Flags().
		StringVar(&jobLogFieldsSince, "since", "1h", "Relative duration to scan (e.g. 5m, 1h); maximum 3d")

	jobsCmd.AddCommand(jobLogsCmd, jobLogFieldsCmd)
	jobRunsCmd.AddCommand(jobRunLogsCmd)
}
