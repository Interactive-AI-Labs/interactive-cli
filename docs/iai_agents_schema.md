## iai agents schema

Display the JSON Schema for agent configuration

### Synopsis

Fetch and display the current JSON Schema for the agent_config block.

Examples:
  iai agents schema
  iai agents schema -o my-org -p my-project

```
iai agents schema [flags]
```

### Options

```
  -h, --help                  help for schema
  -o, --organization string   Organization name
  -p, --project string        Project name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Manage agents

