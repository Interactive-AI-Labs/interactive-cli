## iai dataset-runs list

List dataset runs

### Synopsis

List runs for a given dataset.

```
iai dataset-runs list [flags]
```

### Examples

```
  iai dataset-runs list --dataset-name my-dataset
  iai dataset-runs list --dataset-name my-dataset -o my-org -p my-project
  iai dataset-runs list --dataset-name my-dataset --page 2 --limit 50 --columns name,status
  iai dataset-runs list --dataset-name my-dataset --json
```

### Options

```
      --columns strings       Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml
      --dataset-name string   Dataset name (required)
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Items per page (max 100)
  -o, --organization string   Organization name that owns the project
      --page int              Page number (starts at 1) (default 1)
  -p, --project string        Project name
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

* [iai dataset-runs](iai_dataset-runs.md)	 - Run evaluations against datasets

