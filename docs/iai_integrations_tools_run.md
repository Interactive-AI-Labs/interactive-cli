## iai integrations tools run

Run a tool on a connection

### Synopsis

Invoke a tool exposed by a connection's MCP server and print the result.

Only enabled, server-advertised tools can run. Arguments are a JSON object passed
inline with --args or from a file with --args-file (mutually exclusive). When
omitted, an empty argument object is sent.

Examples:
  iai integrations tools run 3f9c1a2e-... search --args '{"query":"langfuse"}'
  iai integrations tools run 3f9c1a2e-... search --args-file ./args.json

```
iai integrations tools run <connection_id> <tool_name> [flags]
```

### Options

```
      --args string           Tool arguments as an inline JSON object
      --args-file string      Path to a file containing the tool arguments as a JSON object
  -h, --help                  help for run
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connection
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai integrations tools](iai_integrations_tools.md)	 - Inspect and run tools on a connection

