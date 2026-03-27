## iai datasets list

List datasets

### Synopsis

List evaluation datasets with pagination.

```
iai datasets list [flags]
```

### Options

```
      --columns strings       Columns to display (comma-separated)
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Items per page
  -o, --organization string   Organization name that owns the project
      --page int              Page number (starts at 1) (default 1)
  -p, --project string        Project name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai datasets](iai_datasets.md)	 - Manage evaluation datasets

