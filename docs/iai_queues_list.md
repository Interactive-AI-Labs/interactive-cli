## iai queues list

List annotation queues

### Synopsis

List annotation queues with pagination and sorting.

Results are sorted by created_at descending unless --sort-by/--sort-order say otherwise.

```
iai queues list [flags]
```

### Examples

```
  iai queues list
  iai queues list --page 2 --limit 50
  iai queues list --sort-by name --sort-order asc
  iai queues list --sort-by count_pending_items --sort-order desc
  iai queues list --columns id,name,description
  iai queues list -o my-org -p my-project --json
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
      --sort-by string        Sort by field: name, description, created_at, updated_at, count_completed_items, count_pending_items (default: created_at)
      --sort-order string     Sort direction: asc or desc (default: desc)
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

* [iai queues](iai_queues.md)	 - Annotation queues for human review workflows

