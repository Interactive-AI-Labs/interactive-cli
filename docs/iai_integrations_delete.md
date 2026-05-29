## iai integrations delete

Delete an integration connection

### Synopsis

Delete an integration connection and its cached tools. This does not affect the
remote MCP server. Use -f to skip the confirmation prompt.

Examples:
  iai integrations delete 3f9c1a2e-...
  iai integrations delete 3f9c1a2e-... -f

```
iai integrations delete <connection-id> [flags]
```

### Options

```
  -f, --force                 Skip confirmation prompt
  -h, --help                  help for delete
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

