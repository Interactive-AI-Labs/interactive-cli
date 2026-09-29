## iai jobs log-fields

List available fields in structured logs

### Synopsis

Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai jobs logs --fields' to include them in output.

```
iai jobs log-fields <job_name> [flags]
```

### Examples

```
  iai jobs log-fields my-job
  iai jobs log-fields my-job --since 1h
```

### Options

```
  -h, --help           help for log-fields
      --since string   Relative duration to scan (e.g. 5m, 1h); maximum 3d (default "1h")
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

