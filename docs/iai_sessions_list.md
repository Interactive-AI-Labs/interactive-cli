## iai sessions list

List sessions

### Synopsis

List sessions with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.

```
iai sessions list [flags]
```

### Options

```
      --columns strings         Columns to display (comma-separated, default: id,created_at,environment,trace_count,duration_seconds,total_cost,total_tokens)
                                Available: id,created_at,updated_at,environment,user_id,trace_count,duration_seconds,total_cost,input_tokens,output_tokens,total_tokens
      --environment string      Filter by environment
      --from-timestamp string   Filter sessions from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                    help for list
      --json                    Output raw API response as JSON
      --limit int               Items per page
  -o, --organization string     Organization name that owns the project
      --page int                Page number (starts at 1) (default 1)
  -p, --project string          Project name
      --to-timestamp string     Filter sessions to this timestamp (ISO 8601)
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

* [iai sessions](iai_sessions.md)	 - Manage sessions

