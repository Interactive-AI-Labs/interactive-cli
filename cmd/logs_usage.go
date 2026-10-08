package cmd

// The index lifts the "(max …)" notes from these.
const (
	logsMaxWindow  = "72h"
	logsSinceUsage = "Relative duration to look back (max " + logsMaxWindow +
		", e.g. 30m, 1h, 3d, 1w); default 1h; mutually exclusive with --start-time and --end-time"
	logsStartTimeUsage = "Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); " +
		"mutually exclusive with --since; max " + logsMaxWindow + " window"
	logsLimitUsage = "Maximum number of log entries to return (max 5000); defaults to 1000"
)
