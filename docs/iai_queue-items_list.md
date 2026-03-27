## iai queue-items list

List queue items

### Synopsis

List items in an annotation queue.

```
iai queue-items list [flags]
```

### Options

```
      --columns strings       Columns to display (comma-separated)
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Items per page (default 50)
  -o, --organization string   Organization name that owns the project
      --page int              Page number (starts at 1) (default 1)
  -p, --project string        Project name
      --queue-id string       Queue ID (required)
      --status string         Filter by status (PENDING/COMPLETED)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai queue-items](iai_queue-items.md)	 - Manage annotation queue items

