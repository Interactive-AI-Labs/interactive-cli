## iai mcps delete

Delete an mcp

### Synopsis

Remove the mcp's release from the project namespace — its workload (if
internal), credential Secret, and cached tools. Rejected if agents are still
attached, unless -f is also set, in which case the delete proceeds and those
agents keep a dangling reference until it's removed. -f also skips the
confirmation prompt.

Detach it from any attached agent first with 'iai agents update <agent> --detach-mcp <mcp_name>'.

```
iai mcps delete <mcp_name> [flags]
```

### Examples

```
  iai mcps delete my-tool
  iai mcps delete my-tool -f
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for delete
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

