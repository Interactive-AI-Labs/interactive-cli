## iai connectors list

List connectors in a project

### Synopsis

List MCP connectors in a project, showing type, status, tool count, and endpoint.

Examples:
  iai connectors list

```
iai connectors list [flags]
```

### Options

```
  -h, --help                  help for list
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connectors
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

