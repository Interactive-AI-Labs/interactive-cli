## iai dataset-runs list

List dataset runs

### Synopsis

List runs for a given dataset.

```
iai dataset-runs list [flags]
```

### Options

```
      --columns strings       Columns to display (comma-separated)
      --dataset-name string   Dataset name (required)
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Items per page (default 50)
  -o, --organization string   Organization name
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

* [iai dataset-runs](iai_dataset-runs.md)	 - Manage dataset runs

