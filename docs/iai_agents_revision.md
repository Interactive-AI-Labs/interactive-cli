## iai agents revision

Describe a specific revision of an agent

### Synopsis

Show the configuration of a specific past revision of an agent.

Examples:
  iai agents revision my-agent 1
  iai agents revision my-agent 3

```
iai agents revision <agent_name> <revision> [flags]
```

### Options

```
  -h, --help                  help for revision
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

