## iai dataset-items create

Create a dataset item

### Synopsis

Create or upsert an item in a dataset.

```
iai dataset-items create [flags]
```

### Options

```
      --dataset-name string            Dataset name (required)
      --expected-output string         Expected output as JSON string
  -h, --help                           help for create
      --id string                      Explicit item ID (for upsert)
      --input string                   Input as JSON string
      --json                           Output raw API response as JSON
      --metadata-json string           Metadata as JSON object
  -o, --organization string            Organization name
  -p, --project string                 Project name
      --source-observation-id string   Source observation ID
      --source-trace-id string         Source trace ID
      --status string                  Item status (ACTIVE/ARCHIVED)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai dataset-items](iai_dataset-items.md)	 - Manage dataset items

