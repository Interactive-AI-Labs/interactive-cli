## iai connectors delete

Delete a connector

### Synopsis

Delete a connector and its cached tools. Does not affect the remote MCP server.
Use -f to skip confirmation.

Examples:
  iai connectors delete 3f9c1a2e-...
  iai connectors delete 3f9c1a2e-... -f

```
iai connectors delete <connector_id> [flags]
```

### Options

```
  -f, --force                 Skip confirmation prompt
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connector
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai connectors](iai_connectors.md)	 - Manage MCP connectors in a project

