## iai replicas log-fields

List available fields in structured logs

### Synopsis

Scan recent logs and list the extra fields present in structured (JSON) log entries.

Use the reported field names with 'iai replicas logs --fields' to include them in output.

Examples:
  iai replicas log-fields my-service-abc123
  iai replicas log-fields my-service-abc123 --since 1h

```
iai replicas log-fields <replica_name> [flags]
```

### Options

```
  -h, --help                  help for log-fields
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the service
      --since string          Relative duration to scan (e.g. 5m, 1h) (default "1h")
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

