## iai mcps diff

Compare two revisions of an mcp

### Synopsis

Show the config differences between two revisions of an mcp — spec only.
Cached tools change at verify time, not per revision; 'iai mcps tools' shows
what changed since the previous verify.

```
iai mcps diff <mcp_name> <revision_a> <revision_b> [flags]
```

### Examples

```
  iai mcps diff my-tool 1 3
```

### Options

```
  -h, --help   help for diff
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

