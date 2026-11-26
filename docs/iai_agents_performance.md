## iai agents performance

Show traces, sessions, latency and cost per agent

### Synopsis

Show the project's agent table: traces, sessions, latency and cost per agent, and the total cost of the agents shown.

Without --agent, the selected agent is used (see 'iai agents select'); with neither, every agent.

```
iai agents performance [flags]
```

### Examples

```
  iai agents performance
  iai agents performance --agent my-agent --agent other-agent
  iai agents performance --from-timestamp 2026-11-01T00:00:00Z --json
```

### Options

```
      --agent stringArray       Agent name (repeatable); defaults to the selected agent, else every agent
      --from-timestamp string   Show agents from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                    help for performance
      --json                    Output raw API response as JSON
  -o, --organization string     Organization name
  -p, --project string          Project name
      --to-timestamp string     Show agents to this timestamp (ISO 8601, default: now)
      --yaml                    Output raw API response as YAML
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

