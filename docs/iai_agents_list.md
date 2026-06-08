## iai agents list

List agents in a project

### Synopsis

List agents in a specific project.

Examples:
  iai agents list
  iai agents list -p my-project

```
iai agents list [flags]
```

### Options

```
  -h, --help                  help for list
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output raw API response as YAML
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

