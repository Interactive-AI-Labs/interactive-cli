## iai agents catalog

List available agent types and versions

### Synopsis

List agent types available in the catalog.

Without arguments, lists all available agent IDs.
With an agent ID argument, lists available versions for that agent.

Examples:
  iai agents catalog
  iai agents catalog interactive-agent

```
iai agents catalog [agent_id] [flags]
```

### Options

```
  -h, --help   help for catalog
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

