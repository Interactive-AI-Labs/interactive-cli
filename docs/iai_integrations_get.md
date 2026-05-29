## iai integrations get

Show an integration connection and its tools

### Synopsis

Show detailed information about a single integration connection, including the
cached list of tools discovered from the MCP server.

Examples:
  iai integrations get 3f9c1a2e-...

```
iai integrations get <connection_id> [flags]
```

### Options

```
  -h, --help                  help for get
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

