## iai agents overview

Show an agent's performance, cost, usage and errors

### Synopsis

Show an agent's Overview: performance, cost, tokens, usage and errors for a window, with the previous window of the same length.

Without agent_name, the selected agent is used (see 'iai agents select').

```
iai agents overview [agent_name] [flags]
```

### Examples

```
  iai agents overview my-agent
  iai agents overview my-agent --window 7d --json
  iai agents overview my-agent --window custom --from-timestamp 2026-11-01T00:00:00Z --to-timestamp 2026-11-08T00:00:00Z
```

### Options

```
      --from-timestamp string   Start of a custom window (ISO 8601)
  -h, --help                    help for overview
      --json                    Output raw API response as JSON
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name
      --to-timestamp string     End of a custom window (ISO 8601)
      --window string           Window length: 30m, 1h, 6h, 24h, 3d, 7d, 14d, 1mo, 3mo, all, or custom with --from-timestamp and --to-timestamp (default "24h")
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

