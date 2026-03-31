## iai scores list

List scores

### Synopsis

List scores with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.

```
iai scores list [flags]
```

### Options

```
      --columns strings         Columns to display (comma-separated, default: id,name,data_type,value,source,timestamp,trace_id)
                                Available: id,name,data_type,value,source,timestamp,trace_id,observation_id,session_id,environment,config_id,user_id,comment
      --config-id string        Filter by config ID
      --cursor string           Cursor for pagination
      --data-type string        Filter by data type
      --environment string      Filter by environment
      --fields string           Field groups to include (comma-separated)
      --from-timestamp string   Filter scores from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                    help for list
      --json                    Output raw API response as JSON
      --limit int               Items per page
      --max-value float         Maximum score value
      --min-value float         Minimum score value
      --name string             Filter by score name
      --observation-id string   Filter by observation ID
      --operator string         Operator for --value
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name
      --score-id stringArray    Filter by score ID (repeatable)
      --session-id string       Filter by session ID
      --source string           Filter by source
      --to-timestamp string     Filter scores to this timestamp (ISO 8601)
      --trace-id string         Filter by trace ID
      --trace-tag stringArray   Filter by trace tag (repeatable)
      --user-id string          Filter by user ID
      --value string            Exact value filter
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai scores](iai_scores.md)	 - Manage scores

