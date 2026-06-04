## iai connectors list

List connectors in a project

### Synopsis

Show each connector's type, status, tool count, and endpoint in a table.

Examples:
  iai connectors list
  iai connectors list --json

```
iai connectors list [flags]
```

### Options

```
  -h, --help   help for list
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

