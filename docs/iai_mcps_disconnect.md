## iai mcps disconnect

Forget an mcp's stored provider credential

### Synopsis

Remove the provider credential this mcp holds, so it can no longer be used
until someone signs in again.

The connection is shared by the whole project, so this affects every agent and
everyone in it. The mcp itself is kept — use 'iai mcps delete' to remove that.

```
iai mcps disconnect <mcp_name> [flags]
```

### Examples

```
  iai mcps disconnect notion-demo
```

### Options

```
  -h, --help   help for disconnect
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

