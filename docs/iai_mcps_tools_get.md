## iai mcps tools get

List an mcp's cached tools with descriptions

### Synopsis

Show the full cached tool list — name and description. 'iai mcps get' only
shows a count; use this to see the tools themselves.

--revision reads a past helm release revision's snapshot instead of the
current one (see 'iai mcps tools revisions').

```
iai mcps tools get <mcp_name> [flags]
```

### Examples

```
  iai mcps tools get my-tool
  iai mcps tools get my-tool --revision 3
  iai mcps tools get my-tool --json
```

### Options

```
  -h, --help           help for get
      --json           Output raw API response as JSON
      --revision int   Read this past helm release revision's cached tools instead of the current one
      --yaml           Output raw API response as YAML
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

