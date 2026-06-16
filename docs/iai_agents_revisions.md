## iai agents revisions

List revisions of an agent

### Synopsis

Show past revisions of an agent, sorted newest-first.
Up to 50 revisions are retained per agent.

```
iai agents revisions <agent_name> [flags]
```

### Examples

```
  iai agents revisions my-agent
```

### Options

```
  -h, --help                  help for revisions
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

