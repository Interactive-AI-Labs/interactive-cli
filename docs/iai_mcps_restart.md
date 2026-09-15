## iai mcps restart

Restart an mcp in a project

### Synopsis

Restart an mcp in a specific project using the deployment service.

```
iai mcps restart <mcp_name> [flags]
```

### Examples

```
  iai mcps restart my-tool
  iai mcps restart my-tool --project my-project
```

### Options

```
  -h, --help   help for restart
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

