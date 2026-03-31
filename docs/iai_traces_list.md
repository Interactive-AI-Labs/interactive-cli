## iai traces list

List traces

### Synopsis

List traces with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.

Examples:
  iai traces list
  iai traces list --limit 20 --page 2
  iai traces list --name my-trace --user-id user123
  iai traces list --from-timestamp 2025-01-01T00:00:00Z
  iai traces list --order-by timestamp --order desc
  iai traces list --tags tag1 --tags tag2
  iai traces list --model gpt-4 --has-error
  iai traces list --min-cost 0.01 --max-cost 1.0
  iai traces list --level ERROR
  iai traces list --search "my query"
  iai traces list --fields core,io,metrics
  iai traces list --json | jq '.data.traces[].name'
  iai traces list --columns id,name,latency,total_tokens,level

```
iai traces list [flags]
```

### Options

```
      --columns strings           Columns to display (comma-separated, default: id,name,timestamp,latency,cost,tags)
                                  Available: id,name,timestamp,user_id,session_id,release,version,environment,public,latency,cost,tags,observation_count,input_tokens,output_tokens,total_tokens,level
      --environment stringArray   Filter by environment (repeatable)
      --fields string             Field groups to include: core, io, metrics (comma-separated) (default "core,metrics")
      --from-timestamp string     Filter traces from this timestamp (ISO 8601, default: 7 days ago)
      --has-error                 Filter traces with errors
  -h, --help                      help for list
      --json                      Output raw API response as JSON
      --level string              Filter by aggregated level: DEBUG, DEFAULT, WARNING, ERROR
      --limit int                 Items per page
      --max-cost float            Maximum total cost filter
      --max-latency float         Maximum latency filter (seconds)
      --max-tokens int            Maximum total tokens filter
      --min-cost float            Minimum total cost filter
      --min-latency float         Minimum latency filter (seconds)
      --min-tokens int            Minimum total tokens filter
      --model string              Filter by model name
      --name string               Filter by trace name
      --order string              Sort direction: asc or desc (default: desc) (default "desc")
      --order-by string           Order by field: timestamp, latency, cost, name
  -o, --organization string       Organization name that owns the project
      --page int                  Page number (starts at 1) (default 1)
  -p, --project string            Project name
      --release string            Filter by release
      --search string             Search in trace name (max 200 characters)
      --session-id string         Filter by session ID
      --tags stringArray          Filter by tags (repeatable)
      --to-timestamp string       Filter traces to this timestamp (ISO 8601)
      --user-id string            Filter by user ID
      --version string            Filter by version
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

* [iai traces](iai_traces.md)	 - Manage traces

