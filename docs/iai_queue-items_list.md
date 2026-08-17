## iai queue-items list

List queue items

### Synopsis

List items in an annotation queue.

Results are sorted by created_at descending unless --sort-by/--sort-order say otherwise.

```
iai queue-items list [flags]
```

### Examples

```
  iai queue-items list --queue-id queue-123
  iai queue-items list --queue-id queue-123 --status PENDING
  iai queue-items list --queue-id queue-123 --page 2 --limit 50
  iai queue-items list --queue-id queue-123 --sort-by created_at --sort-order asc
  iai queue-items list --queue-id queue-123 --sort-by status --sort-order desc
  iai queue-items list --queue-id queue-123 --json
```

### Options

```
      --columns strings       Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Items per page (max 100)
  -o, --organization string   Organization name that owns the project
      --page int              Page number (starts at 1) (default 1)
  -p, --project string        Project name
      --queue-id string       Queue ID (required)
      --sort-by string        Sort by field: created_at, completed_at, status, object_id, updated_at (default: created_at)
      --sort-order string     Sort direction: asc or desc (default: desc)
      --status string         Filter by status (PENDING/COMPLETED)
      --yaml                  Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai queue-items](iai_queue-items.md)	 - Manage items in annotation queues

