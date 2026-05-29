## iai integrations verify

Re-verify a connection and refresh its tools

### Synopsis

Re-dial the MCP server for a connection (initialize + list tools) and refresh the
cached tool list. Reports the connection status and, on failure, the error class
and message.

Examples:
  iai integrations verify 3f9c1a2e-...

```
iai integrations verify <connection-id> [flags]
```

### Options

```
  -h, --help                  help for verify
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connection
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai integrations](iai_integrations.md)	 - MCP integration connections for a project

