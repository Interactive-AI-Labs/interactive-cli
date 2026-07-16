## iai mcps get

Show mcp details, verify state, and cached tools

### Synopsis

Show the mcp's record (type, external URL, catalog origin) and its latest
verify result — a tool count, not the tool list itself (see 'iai mcps tools').

```
iai mcps get <mcp_name> [flags]
```

### Examples

```
  iai mcps get my-tool
  iai mcps get my-tool --json
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
  -p, --project string               Project name that owns the mcps
```

### SEE ALSO

* [iai mcps](iai_mcps.md)	 - Deploy and manage MCP servers

