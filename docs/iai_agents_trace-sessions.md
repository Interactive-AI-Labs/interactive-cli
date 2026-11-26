## iai agents trace-sessions

List an agent's trace sessions

### Synopsis

List an agent's trace sessions.

Without agent_name, the selected agent is used (see 'iai agents select').

```
iai agents trace-sessions [agent_name] [flags]
```

### Examples

```
  iai agents trace-sessions my-agent
  iai agents trace-sessions my-agent --limit 20 --json
```

### Options

```
      --columns strings         Columns to display for table output only (comma-separated, default: id,created_at,environment,trace_count,duration_seconds,total_cost,total_tokens). Cannot be used with --json or --yaml.
                                Available: id,created_at,updated_at,environment,user_id,trace_count,duration_seconds,total_cost,input_tokens,output_tokens,total_tokens
      --environment string      Filter by environment
      --from-timestamp string   Filter sessions from this timestamp (ISO 8601, default: 7 days ago)
  -h, --help                    help for trace-sessions
      --json                    Output raw API response as JSON
      --limit int               Items per page (max 100)
  -o, --organization string     Organization name that owns the project
      --page int                Page number (starts at 1) (default 1)
  -p, --project string          Project name
      --to-timestamp string     Filter sessions to this timestamp (ISO 8601)
      --yaml                    Output raw API response as YAML
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

