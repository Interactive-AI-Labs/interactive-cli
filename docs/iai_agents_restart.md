## iai agents restart

Restart an agent in a project

### Synopsis

Restart an agent in a specific project.

Examples:
  iai agents restart my-agent

```
iai agents restart <agent_name> [flags]
```

### Options

```
  -h, --help                  help for restart
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

