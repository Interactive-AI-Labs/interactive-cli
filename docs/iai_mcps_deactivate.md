## iai mcps deactivate

Deactivate an internal mcp in a project

### Synopsis

Deactivate an internal mcp, stopping all running instances. The current
configuration is preserved and will be restored when the mcp is activated again.
External mcps run no workload and cannot be deactivated.

```
iai mcps deactivate <mcp_name> [flags]
```

### Examples

```
  iai mcps deactivate my-mcp
  iai mcps deactivate my-mcp --project my-project
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

