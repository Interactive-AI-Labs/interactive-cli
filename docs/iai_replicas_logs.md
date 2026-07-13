## iai replicas logs

Show logs for a specific replica

### Synopsis

Show logs for a specific replica in a project.

Returns up to 1000 log entries in chronological order by default; use
--limit to request up to 5000.

Structured (JSON) logs are automatically formatted: the level and message
fields are extracted and displayed as "LEVEL message". Use --fields or
--all-fields to include additional top-level fields after the message. Use
--raw for exact server JSON, or --decode to decode embedded JSON strings into
nested JSON values.

```
iai replicas logs <replica_name> [flags]
```

### Examples

```
  iai replicas logs my-service-abc123
  iai replicas logs my-service-abc123 --follow
  iai replicas logs my-service-abc123 --since 30m --fields logger,pid
  iai replicas logs my-service-abc123 --timestamps
  iai replicas logs my-service-abc123 --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z
```

### Options

```
      --all-fields            Show all extra top-level fields from structured (JSON) logs after the message
      --decode                Decode embedded JSON strings into nested JSON values; outputs raw JSON
      --end-time string       Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow
      --fields strings        Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON
  -f, --follow                Stream new log entries as they arrive; mutually exclusive with --end-time
  -h, --help                  help for logs
      --limit int             Maximum number of log entries to return (1-5000); defaults to 1000
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the service
      --raw                   Output exact server JSON lines without formatting
      --since string          Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time
      --start-time string     Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window
      --timestamps            Include platform log timestamps
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai replicas](iai_replicas.md)	 - Inspect service replicas

