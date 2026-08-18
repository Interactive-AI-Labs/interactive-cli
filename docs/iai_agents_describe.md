## iai agents describe

Describe an agent in detail

### Synopsis

Show detailed information about a specific agent including its configuration.

Use --revision to view a specific past revision instead of the current state.
Past revision output includes server-recorded actor and source attribution when available.

```
iai agents describe <agent_name> [flags]
```

### Examples

```
  iai agents describe my-agent
  iai agents describe my-agent --revision 3
  iai agents describe my-agent --yaml
```

### Options

```
  -h, --help                  help for describe
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --revision int          Show a specific past revision instead of the current state
  -w, --watch                 Poll and refresh every 2s until interrupted
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

