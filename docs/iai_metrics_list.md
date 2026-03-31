## iai metrics list

List observability metrics

### Synopsis

List observability metrics with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.
Use --daily to get metrics aggregated by day (default).

Examples:
  iai metrics list --daily
  iai metrics list --daily --from-timestamp 2025-01-01T00:00:00Z
  iai metrics list --daily --trace-name my-trace --show-models
  iai metrics list --daily --json | jq '.data'

```
iai metrics list [flags]
```

### Options

```
      --columns strings         Columns to display (comma-separated, default: date,count_traces,count_observations,total_cost)
                                Available: date,count_traces,count_observations,total_cost,total_tokens
      --daily                   Aggregate metrics by day (default true)
      --environment string      Filter by environment
      --from-timestamp string   Filter metrics from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                    help for list
      --json                    Output raw API response as JSON
      --limit int               Items per page
  -o, --organization string     Organization name that owns the project
      --page int                Page number (starts at 1) (default 1)
  -p, --project string          Project name
      --show-models             Show per-model breakdown
      --tags stringArray        Filter by tags (repeatable)
      --to-timestamp string     Filter metrics to this timestamp (ISO 8601)
      --trace-name string       Filter by trace name
      --user-id string          Filter by user ID
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

* [iai metrics](iai_metrics.md)	 - Manage observability metrics

