## iai dataset-items create

Create a dataset item

### Synopsis

Create or upsert an item in a dataset.

```
iai dataset-items create [flags]
```

### Examples

```
  iai dataset-items create --dataset-name my-dataset --input '{"question":"2+2?"}'
  iai dataset-items create --dataset-name my-dataset --name defensores-37-pdf-excel-v2 --input '{"q":"hi"}'
  iai dataset-items create --dataset-name my-dataset --input '{"q":"hi"}' --status ARCHIVED
  iai dataset-items create --dataset-name my-dataset --input '{"question":"2+2?"}' --expected-output '{"answer":"4"}'
  iai dataset-items create --dataset-name my-dataset --input '{"q":"hi"}' --metadata-json '{"source":"manual"}'
  iai dataset-items create --dataset-name my-dataset --input '{"q":"hi"}' --source-trace-id trace-123 --json
```

### Options

```
      --dataset-name string            Dataset name (required)
      --expected-output string         Expected output as JSON string
  -h, --help                           help for create
      --id string                      Explicit item ID; upserts an existing item (needs datasets:update)
      --input string                   Input as JSON string
      --json                           Output raw API response as JSON
      --metadata-json string           Metadata as JSON object
      --name string                    Item name
  -o, --organization string            Organization name that owns the project
  -p, --project string                 Project name
      --source-observation-id string   Source observation ID
      --source-trace-id string         Source trace ID
      --status string                  Item status; ARCHIVED needs datasets:update (ACTIVE|ARCHIVED)
      --yaml                           Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai dataset-items](iai_dataset-items.md)	 - Manage items in evaluation datasets

