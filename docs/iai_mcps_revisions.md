## iai mcps revisions

List revisions of an mcp

### Synopsis

Show past revisions of an mcp, sorted newest-first. Up to 50 revisions are
retained per mcp. Every spec change — update, credential rotation, agent
attach/detach — creates a revision.

```
iai mcps revisions <mcp_name> [flags]
```

### Examples

```
  iai mcps revisions my-tool
```

### Options

```
  -h, --help   help for revisions
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

