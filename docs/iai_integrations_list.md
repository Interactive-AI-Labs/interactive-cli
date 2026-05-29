## iai integrations list

List integration connections in a project

### Synopsis

List the MCP integration connections in a project, showing each connection's
type, status, tool count, and endpoint.

Examples:
  iai integrations list

```
iai integrations list [flags]
```

### Options

```
  -h, --help                  help for list
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connections
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

