## iai mcps activate

Activate a deactivated internal mcp in a project

### Synopsis

Activate a deactivated internal mcp, restoring it to its previous
configuration. External mcps run no workload and cannot be activated.

```
iai mcps activate <mcp_name> [flags]
```

### Examples

```
  iai mcps activate my-mcp
  iai mcps activate my-mcp --project my-project
```

### Options

```
  -h, --help   help for activate
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

