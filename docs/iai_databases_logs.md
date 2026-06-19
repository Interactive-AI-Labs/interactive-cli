## iai databases logs

Show logs for a database

### Synopsis

Show logs for a database in a project.

Returns up to 5000 log entries in chronological order. Default lookback is 1h.

Structured (JSON) logs are automatically formatted: the level and message are
extracted and displayed. PostgreSQL-style logs use a "record" envelope — the
severity and message are extracted from it automatically. Use --fields record
to see the nested PostgreSQL details; --all-fields includes extra top-level
fields only. Use --raw for exact server JSON, or --decode to decode embedded
JSON strings into nested JSON values.

```
iai databases logs <database_name> [flags]
```

### Examples

```
  iai databases logs my-db
  iai databases logs my-db --follow
  iai databases logs my-db --since 30m
  iai databases logs my-db --timestamps
  iai databases logs my-db --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z
```

### Options

```
      --all-fields            Show all extra top-level fields from structured (JSON) logs after the message
      --decode                Decode embedded JSON strings into nested JSON values; outputs raw JSON
      --end-time string       Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow
      --fields strings        Additional fields to show after the message for structured (JSON) logs (e.g. --fields record); ignored for plain-text logs; use --raw for exact server JSON
  -f, --follow                Stream new log entries as they arrive; mutually exclusive with --end-time
  -h, --help                  help for logs
  -o, --organization string   Organization name
  -p, --project string        Project name
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

* [iai databases](iai_databases.md)	 - PostgreSQL instances with extension support, including pgvector

