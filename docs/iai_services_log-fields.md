## iai services log-fields

List available fields in structured logs

### Synopsis

Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai services logs --fields' to include them in output.

```
iai services log-fields <service_name> [flags]
```

### Examples

```
  iai services log-fields my-service
  iai services log-fields my-service --since 1h
```

### Options

```
  -h, --help                  help for log-fields
  -o, --organization string   Organization name
  -p, --project string        Project name
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

* [iai services](iai_services.md)	 - Deploy and manage HTTP services

