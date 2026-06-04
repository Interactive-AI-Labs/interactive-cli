## iai agents schema

Display the JSON Schema for agent configuration

### Synopsis

Fetch and display the JSON Schema for the agent_config block.

Defaults to the latest schema version. Use --schema-version to fetch a specific
version (run 'iai agents compatibility-matrix' to see available versions).

Use --json or --yaml for structured schema output.

Examples:
  iai agents schema
  iai agents schema --schema-version 2.1.0
  iai agents schema --json

```
iai agents schema [flags]
```

### Options

```
  -h, --help                    help for schema
      --json                    Output schema response as JSON
      --schema-version string   Schema version to fetch (defaults to latest stable)
      --yaml                    Output schema response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools

