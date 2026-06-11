## iai score-configs list

List score configs

### Synopsis

List scoring configurations with pagination.

```
iai score-configs list [flags]
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

* [iai score-configs](iai_score-configs.md)	 - Define scoring schemas for evaluation

