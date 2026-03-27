## iai run-items create

Create a run item

### Synopsis

Create a new dataset run item linking a trace/observation to a dataset item.

This command requires API key authentication.

```
iai run-items create [flags]
```

### Options

```
      --dataset-item-id string   Dataset item ID (required)
  -h, --help                     help for create
      --json                     Output raw API response as JSON
      --metadata-json string     Metadata as JSON object
      --observation-id string    Observation ID
  -o, --organization string      Organization name
  -p, --project string           Project name
      --run-description string   Run description
      --run-name string          Run name (required)
      --trace-id string          Trace ID
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

