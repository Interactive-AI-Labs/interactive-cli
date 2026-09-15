## iai mcps logs

Show logs for an mcp

### Synopsis

Show logs for all replicas of an mcp in a project.

Returns up to 1000 log entries in chronological order by default; use
--limit to request up to 5000.

Structured (JSON) logs are automatically formatted: the level and message
fields are extracted and displayed as "LEVEL message". Use --fields or
--all-fields to include additional top-level fields after the message. Use
--raw for exact server JSON, or --decode to decode embedded JSON strings into
nested JSON values.

Use --summary for JSON severity counts over the entire time window, without
the log-entry limit. Unlabeled lines contribute to total only.

```
iai mcps logs <mcp_name> [flags]
```

### Examples

```
  iai mcps logs my-tool
  iai mcps logs my-tool --follow
  iai mcps logs my-tool --since 3h
  iai mcps logs my-tool --timestamps
  iai mcps logs my-tool --fields logger,pid
  iai mcps logs my-tool --since 30m --level error --message 'timeout|failed'
  iai mcps logs my-tool --summary --since 3h
  iai mcps logs my-tool --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z
```

### Options

```
      --all-fields          Show all extra top-level fields from structured (JSON) logs after the message
      --decode              Decode embedded JSON strings into nested JSON values; outputs raw JSON
      --end-time string     Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow
      --fields strings      Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON
  -f, --follow              Stream new log entries as they arrive; mutually exclusive with --end-time
  -h, --help                help for logs
      --level string        JSON log severity: debug, info, warn, or error
      --limit int           Maximum number of log entries to return (1-5000); defaults to 1000
      --message string      Case-insensitive RE2 regular expression matched against the log line
      --raw                 Output exact server JSON lines without formatting
      --since string        Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time
      --start-time string   Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window
      --summary             Output JSON severity counts instead of log entries
      --timestamps          Include platform log timestamps
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the mcps
```

### SEE ALSO

* [iai mcps](iai_mcps.md)	 - Deploy and manage MCP servers

