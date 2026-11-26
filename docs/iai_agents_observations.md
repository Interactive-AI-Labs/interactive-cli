## iai agents observations

List an agent's observations

```
iai agents observations <agent_name> [flags]
```

### Examples

```
  iai agents observations my-agent
  iai agents observations my-agent --limit 20 --json
```

### Options

```
      --columns strings                Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml
                                       With --trace-id default: id,type,name,model,latency_ms,total_cost,total_tokens
                                       Without --trace-id default: id,trace_id,type,name,model,latency_ms,total_cost,total_tokens
      --cursor string                  Cursor for pagination
      --environment string             Filter by environment
      --fields string                  Field groups to include (comma-separated)
      --from-timestamp string          Filter observations from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                           help for observations
      --json                           Output raw API response as JSON
      --level string                   Filter by level
      --limit int                      Items per page (max 100)
      --model string                   Filter by model
      --name string                    Filter by observation name
  -o, --organization string            Organization name that owns the project
      --parent-observation-id string   Filter by parent observation ID
  -p, --project string                 Project name
      --to-timestamp string            Filter observations to this timestamp (ISO 8601)
      --type string                    Filter by observation type
      --user-id string                 Filter by user ID
      --version string                 Filter by version
      --yaml                           Output raw API response as YAML
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

