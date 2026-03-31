## iai dataset-items list

List dataset items

### Synopsis

List items in a dataset with optional filters.

```
iai dataset-items list [flags]
```

### Options

```
      --columns strings                Columns to display (comma-separated)
      --dataset-name string            Dataset name (required)
  -h, --help                           help for list
      --json                           Output raw API response as JSON
      --limit int                      Items per page
  -o, --organization string            Organization name that owns the project
      --page int                       Page number (starts at 1) (default 1)
  -p, --project string                 Project name
      --source-observation-id string   Filter by source observation ID
      --source-trace-id string         Filter by source trace ID
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

* [iai dataset-items](iai_dataset-items.md)	 - Manage dataset items

