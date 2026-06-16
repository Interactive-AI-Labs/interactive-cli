## iai connectors get

Show a connector and its tools

### Synopsis

Print a connector's full configuration and status alongside the cached list of
tools discovered from the MCP server.

```
iai connectors get <connector_id> [flags]
```

### Examples

```
  iai connectors get 3f9c1a2e-...
  iai connectors get 3f9c1a2e-... --json
```

### Options

```
  -h, --help   help for get
      --json   Output raw API response as JSON
      --yaml   Output raw API response as YAML
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

