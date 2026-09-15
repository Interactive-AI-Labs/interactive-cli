## iai mcps logs

Show logs for an MCP

### Synopsis

Fetch or follow logs produced by an internal MCP's replicas through the deployment operator.
External provider logs are not collected. Requests query retained logs by MCP name.

Returns up to 1000 entries by default; --limit accepts up to 5000. Structured
logs are formatted as "LEVEL message". Use --fields or --all-fields to include
extra fields, --raw for server JSON lines, or --decode for decoded JSON.

Use --summary for JSON severity counts over the entire time window, without
the log-entry limit. Unlabeled lines contribute to total only.

```
iai mcps logs <mcp_name> [flags]
```

### Examples

```
  iai mcps logs my-tool --follow
  iai mcps logs my-tool --since 30m --level error --message 'timeout|failed'
  iai mcps logs my-tool --fields logger,pid --timestamps
  iai mcps logs my-tool --summary --since 3h
  iai mcps logs my-tool --from-timestamp 2026-01-01T00:00:00Z --to-timestamp 2026-01-01T01:00:00Z
```

### Options

```
      --all-fields              Display all extra structured log fields
      --decode                  Decode embedded JSON strings; output JSON
      --fields strings          Extra structured log fields to display (e.g. logger,pid)
  -f, --follow                  Stream new log entries
      --from-timestamp string   RFC3339 start timestamp; mutually exclusive with --since
  -h, --help                    help for logs
      --level string            JSON log severity: debug, info, warn, or error
      --limit int               Maximum log entries (1-5000); defaults to 1000
      --message string          Case-insensitive RE2 regular expression matched against the log line
      --raw                     Output exact server JSON lines without formatting
      --since string            Relative lookback (e.g. 30m, 1h, 3d); default 1h; maximum 72h
      --summary                 Output JSON severity counts instead of log entries
      --timestamps              Include platform log timestamps
      --to-timestamp string     RFC3339 end timestamp; requires --from-timestamp; mutually exclusive with --since and --follow
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

