## iai traces get

Get a specific trace

### Synopsis

Get detailed information about a specific trace.

Uses the platform API with dual authentication (API key or session).

Examples:
  iai traces get abc123
  iai traces get abc123 --fields core,io,metrics
  iai traces get abc123 --json | jq '.data.trace'

```
iai traces get <trace-id> [flags]
```

### Options

```
      --fields string         Field groups to include: core, io, metrics (comma-separated) (default "core,io,metrics")
  -h, --help                  help for get
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
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

