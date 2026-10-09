package cmd

import "github.com/spf13/cobra"

const (
	logsSinceUsage = "Relative duration to look back (max 72h, e.g. 30m, 1h, 3d, 1w); default 1h; " +
		"mutually exclusive with --start-time and --end-time"
	logsStartTimeUsage = "Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); " +
		"mutually exclusive with --since; max 72h window"
	logsLimitUsage = "Maximum number of log entries to return (1-5000); defaults to 1000"
)

func indexLogsLimits(c *cobra.Command) {
	indexLimit(c, "since", "max 72h")
	indexLimit(c, "start-time", "max 72h window")
	indexLimit(c, "limit", "1-5000")
}
