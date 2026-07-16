## iai mcps tools

List an mcp's cached tools with descriptions

### Synopsis

Show the full cached tool list — name and description. 'iai mcps get' only
shows a count; use this to see the tools themselves.

```
iai mcps tools <mcp_name> [flags]
```

### Examples

```
  iai mcps tools my-tool
  iai mcps tools my-tool --json
```

### Options

```
  -h, --help   help for tools
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
  -p, --project string               Project name that owns the mcps
```

### SEE ALSO

* [iai mcps](iai_mcps.md)	 - Deploy and manage MCP servers

