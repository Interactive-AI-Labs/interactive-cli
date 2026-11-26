## iai agents analytics

Show an agent's CPU, memory, requests, status codes and latency

### Synopsis

Show an agent's deployment analytics for a time range: the latest and peak value of each CPU, memory, requests, status code and latency series. --json and --yaml print the five raw responses keyed cpu_usage, memory_usage, status_codes, requests_per_domain and latency.

Without agent_name, the selected agent is used (see 'iai agents select').

```
iai agents analytics [agent_name] [flags]
```

### Examples

```
  iai agents analytics my-agent
  iai agents analytics my-agent --time-range 24h
  iai agents analytics my-agent --time-range 6h --step 5m --json
```

### Options

```
  -h, --help                  help for analytics
      --json                  Output raw API responses as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --step string           Resolution, e.g. 60s or 5m (defaults by time range)
      --time-range string     Time range: 5m, 15m, 1h, 6h, 24h or 7d (default 1h)
      --yaml                  Output raw API responses as YAML
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

