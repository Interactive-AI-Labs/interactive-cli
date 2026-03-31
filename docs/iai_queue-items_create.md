## iai queue-items create

Create a queue item

### Synopsis

Add an item to an annotation queue.

This command requires API key authentication.

```
iai queue-items create [flags]
```

### Options

```
  -h, --help                  help for create
      --json                  Output raw API response as JSON
      --object-id string      Object ID (required)
      --object-type string    Object type: TRACE or OBSERVATION (required)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --queue-id string       Queue ID (required)
      --status string         Status (PENDING/COMPLETED)
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

* [iai queue-items](iai_queue-items.md)	 - Manage annotation queue items

