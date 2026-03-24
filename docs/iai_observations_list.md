## iai observations list

List observations for a trace

### Synopsis

List observations for a trace, showing individual LLM calls, spans, and events.

Uses the platform API with dual authentication (API key or session).

Examples:
  iai observations list --trace-id abc123
  iai observations list --trace-id abc123 --include-io
  iai observations list --trace-id abc123 --columns id,type,name,model,latency_ms
  iai observations list --trace-id abc123 --json | jq '.data.observations[].name'

```
iai observations list [flags]
```

### Options

```
      --columns strings       Columns to display (comma-separated, default: id,type,name,model,latency_ms,total_cost,total_tokens)
                              Available: id,trace_id,type,name,start_time,end_time,parent_observation_id,level,status_message,model,input_tokens,output_tokens,total_tokens,total_cost,latency_ms
  -h, --help                  help for list
      --include-io            Include input/output/metadata in response
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --trace-id string       Trace ID to list observations for (required)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai observations](iai_observations.md)	 - Manage observations

