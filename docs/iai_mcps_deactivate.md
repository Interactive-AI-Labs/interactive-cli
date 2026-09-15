## iai mcps deactivate

Stop an MCP's replicas while preserving its configuration

```
iai mcps deactivate <mcp_name> [flags]
```

### Examples

```
  iai mcps deactivate my-tool
```

### Options

```
  -h, --help   help for deactivate
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

