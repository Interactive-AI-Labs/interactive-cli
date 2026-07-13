## iai mcps tools diff

Compare tool sets between two revisions of an mcp

### Synopsis

Show the differences in cached tools between two past revisions of an mcp.

```
iai mcps tools diff <mcp_name> <revision_a> <revision_b> [flags]
```

### Examples

```
  iai mcps tools diff my-tool 1 3
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

* [iai mcps tools](iai_mcps_tools.md)	 - Inspect an mcp's cached tools, current or past

