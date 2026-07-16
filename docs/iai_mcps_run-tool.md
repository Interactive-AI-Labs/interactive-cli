## iai mcps run-tool

Run a tool on an mcp

### Synopsis

Call one of an mcp's tools and print the result.

Pass arguments as a JSON object with --args or --args-file (mutually exclusive);
omit both to send an empty object. Works for external mcps from anywhere;
internal mcps need the in-cluster operator.

```
iai mcps run-tool <mcp_name> <tool_name> [flags]
```

### Examples

```
  iai mcps run-tool github search_repositories --args '{"query":"interactiveai"}'
  iai mcps run-tool github search_repositories --args-file ./args.json
```

### Options

```
      --args string        Tool arguments as an inline JSON object
      --args-file string   Path to a file containing the tool arguments as a JSON object
  -h, --help               help for run-tool
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

