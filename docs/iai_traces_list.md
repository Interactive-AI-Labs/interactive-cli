## iai traces list

List traces

### Synopsis

List traces with optional filters.

Examples:
  iai traces list
  iai traces list --limit 20 --page 2
  iai traces list --name my-trace --user-id user123
  iai traces list --from-timestamp 2025-01-01T00:00:00Z
  iai traces list --order-by timestamp.desc
  iai traces list --tags tag1 --tags tag2

```
iai traces list [flags]
```

### Options

```
      --columns strings           Columns to display (default: id,name,timestamp,latency,cost,tags)
                                  Available: id,name,timestamp,user_id,session_id,release,version,environment,public,latency,cost,tags
      --environment stringArray   Filter by environment (repeatable)
      --from-timestamp string     Filter traces from this timestamp (ISO 8601)
  -h, --help                      help for list
      --limit int                 Items per page
      --name string               Filter by trace name
      --order-by string           Order by field.direction (e.g. timestamp.desc)
      --page int                  Page number (starts at 1)
      --release string            Filter by release
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
```

### SEE ALSO

* [iai traces](iai_traces.md)	 - Manage traces

