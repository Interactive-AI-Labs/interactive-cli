## iai jobs logs

Show logs for a job

### Synopsis

Show logs across a job's runs with 8-character run ID prefixes.
Log retention is independent of run history.

```
iai jobs logs <job_name> [flags]
```

### Examples

```
  iai jobs logs my-job
  iai jobs logs my-job --follow
  iai jobs logs my-job --since 3h
  iai jobs logs my-job --timestamps
  iai jobs logs my-job --fields logger,pid
```

### Options

```
      --all-fields          Show all extra top-level fields from structured (JSON) logs after the message
      --decode              Decode embedded JSON strings into nested JSON values; outputs raw JSON
      --end-time string     Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow
      --fields strings      Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON
  -f, --follow              Stream new entries for up to 10 minutes; reconnect to continue; cannot combine with --end-time
  -h, --help                help for logs
      --limit int           Maximum number of log entries to return (1-5000); defaults to 1000; with --follow, limits only the initial batch
      --raw                 Output exact server JSON lines without formatting
      --since string        Relative duration to look back (max 72h, e.g. 30m, 1h, 3d); default 1h; mutually exclusive with --start-time and --end-time
      --start-time string   Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window
      --timestamps          Include platform log timestamps
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name
```

### SEE ALSO

* [iai jobs](iai_jobs.md)	 - Manage saved jobs and executions

