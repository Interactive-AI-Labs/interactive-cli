## iai agents diff

Compare two revisions of an agent

### Synopsis

Show the differences between two revisions of an agent.

Examples:
  iai agents diff my-agent 1 3

```
iai agents diff <agent_name> <revision_a> <revision_b> [flags]
```

### Options

```
  -h, --help                  help for diff
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

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools

