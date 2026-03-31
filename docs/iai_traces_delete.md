## iai traces delete

Delete one or more traces

### Synopsis

Delete a single trace or bulk delete multiple traces.

This command currently requires API key authentication.

Examples:
  iai traces delete trace-123
  iai traces delete --ids trace-1,trace-2
  iai traces delete --ids trace-1 --ids trace-2 -f

```
iai traces delete [trace-id] [flags]
```

### Options

```
  -f, --force                 Skip bulk delete confirmation
  -h, --help                  help for delete
      --ids strings           Trace IDs to delete (comma-separated or repeatable)
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

