## iai dataset-runs get

Get a dataset run

### Synopsis

Get detailed information about a specific dataset run.

```
iai dataset-runs get <run-name> [flags]
```

### Options

```
      --dataset-name string   Dataset name (required)
  -h, --help                  help for get
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
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

