package cmd

const (
	logsSinceUsage = "Relative duration to look back (max 72h, e.g. 30m, 1h, 3d, 1w); default 1h; " +
		"mutually exclusive with --start-time and --end-time"
	logsLimitUsage = "Maximum number of log entries to return (max 5000); defaults to 1000"
)
