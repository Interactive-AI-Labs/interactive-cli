## iai queues create

Create an annotation queue

### Synopsis

Create a new annotation queue.

This command requires API key authentication.

```
iai queues create <name> [flags]
```

### Options

```
      --description string         Queue description
  -h, --help                       help for create
      --json                       Output raw API response as JSON
  -o, --organization string        Organization name
  -p, --project string             Project name
      --score-config-ids strings   Score config IDs (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai queues](iai_queues.md)	 - Manage annotation queues

