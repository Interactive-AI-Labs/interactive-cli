## iai traces diff

Compare two turns and show where their decision paths diverge

### Synopsis

Compare two traces side by side: routine activations, tools called, and the
per-iteration journey decision path. Highlights the iteration where the agents
selected different routine follow-ups — i.e. where their behavior diverged.

Uses the platform API with dual authentication (API key or session).

```
iai traces diff <trace-id-a> <trace-id-b> [flags]
```

### Examples

```
  iai traces diff abc123 def456
  iai traces diff abc123 def456 --json | jq '.journey'
```

### Options

```
  -h, --help                  help for diff
      --json                  Output the diff as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --yaml                  Output the diff as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai traces](iai_traces.md)	 - Browse agent decision traces with full attribution

