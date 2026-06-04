## iai connectors run-tool

Run a tool on a connector

### Synopsis

Call one of a connector's enabled tools and print the result it returns.

Pass arguments as a JSON object with --args or --args-file (mutually exclusive);
omit both to send an empty object.

Examples:
  iai connectors run-tool 3f9c1a2e-... search --args '{"query":"langfuse"}'
  iai connectors run-tool 3f9c1a2e-... search --args-file ./args.json

```
iai connectors run-tool <connector_id> <tool_name> [flags]
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
  -p, --project string               Project name that owns the connectors
```

### SEE ALSO

* [iai connectors](iai_connectors.md)	 - Manage MCP connectors in a project

