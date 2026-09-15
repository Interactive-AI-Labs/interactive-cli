## iai mcps log-fields

List available fields in structured logs

### Synopsis

Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai mcps logs --fields' to include them in output.

```
iai mcps log-fields <mcp_name> [flags]
```

### Examples

```
  iai mcps log-fields my-tool
  iai mcps log-fields my-tool --since 1h
```

### Options

```
  -h, --help           help for log-fields
      --since string   Relative duration to scan (e.g. 5m, 1h) (default "1h")
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

