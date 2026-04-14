## iai agents describe

Describe an agent in detail

### Synopsis

Show detailed information about a specific agent including its configuration.

Examples:
  iai agents describe my-agent

```
iai agents describe <agent_name> [flags]
```

### Options

```
  -h, --help                  help for describe
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

