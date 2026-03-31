## iai queues assign

Assign a user to a queue

### Synopsis

Assign a user to an annotation queue.

This command requires API key authentication.

```
iai queues assign <queue-id> [flags]
```

### Options

```
  -h, --help                  help for assign
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --user-id string        User ID to assign (required)
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

* [iai queues](iai_queues.md)	 - Manage annotation queues

