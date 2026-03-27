## iai run-items list

List run items

### Synopsis

List dataset run items. Requires at least one of --run-name or --dataset-name.

```
iai run-items list [flags]
```

### Options

```
      --columns strings       Columns to display (comma-separated)
      --dataset-name string   Filter by dataset name
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Items per page (default 50)
  -o, --organization string   Organization name
      --page int              Page number (starts at 1) (default 1)
  -p, --project string        Project name
      --run-name string       Filter by run name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai run-items](iai_run-items.md)	 - Manage dataset run items

