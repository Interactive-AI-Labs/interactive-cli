## iai connectors verify

Re-verify a connector and refresh its tools

### Synopsis

Re-dial the MCP server for a connector (initialize + list tools) and refresh the
cached tool list. Reports the status and, on failure, the error class and message.

Examples:
  iai connectors verify 3f9c1a2e-...
  iai connectors verify 3f9c1a2e-... --json

```
iai connectors verify <connector_id> [flags]
```

### Options

```
  -h, --help   help for verify
      --json   Output raw API response as JSON
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the connectors
```

### SEE ALSO

* [iai connectors](iai_connectors.md)	 - Manage MCP connectors in a project

