## iai observations list

List observations

### Synopsis

List observations for a specific trace or search across traces with filters.

When --trace-id is provided, lists observations within that trace.
Without --trace-id, searches observations across all traces with optional filters.

Uses the platform API with dual authentication (API key or session).

Examples:
  # List observations for a specific trace
  iai observations list --trace-id abc123
  iai observations list --trace-id abc123 --include-io
  iai observations list --trace-id abc123 --columns id,type,name,model,latency_ms

  # Search observations across traces
  iai observations list --type GENERATION --model gpt-4
  iai observations list --from-timestamp 2025-01-01T00:00:00Z --name my-span
  iai observations list --json | jq '.data'

```
iai observations list [flags]
```

### Options

```
      --columns strings                Columns to display (comma-separated)
                                       With --trace-id default: id,type,name,model,latency_ms,total_cost,total_tokens
                                       Without --trace-id default: id,trace_id,type,name,model,latency_ms,total_cost,total_tokens
      --cursor string                  Cursor for pagination
      --environment string             Filter by environment
      --fields string                  Field groups to include (comma-separated)
      --from-timestamp string          Filter observations from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                           help for list
      --include-io                     Include input/output/metadata in response (only with --trace-id)
      --json                           Output raw API response as JSON
      --level string                   Filter by level
      --limit int                      Items per page
      --model string                   Filter by model
      --name string                    Filter by observation name
  -o, --organization string            Organization name that owns the project
      --parent-observation-id string   Filter by parent observation ID
  -p, --project string                 Project name
      --to-timestamp string            Filter observations to this timestamp (ISO 8601)
      --trace-id string                Trace ID to list observations for (scopes to a single trace)
      --type string                    Filter by observation type
      --user-id string                 Filter by user ID
      --version string                 Filter by version
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai observations](iai_observations.md)	 - Manage observations

