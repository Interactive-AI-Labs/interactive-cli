## iai dataset-items update

Update a dataset item

### Synopsis

Change the name, status or metadata of a dataset item. Flags you leave out keep their current value.
--name "" removes the name. --metadata-json replaces the metadata and keeps the name.

```
iai dataset-items update <id> [flags]
```

### Examples

```
  iai dataset-items update item-123 --name defensores-37-pdf-excel-v2
  iai dataset-items update item-123 --status ARCHIVED
  iai dataset-items update item-123 --status ACTIVE --metadata-json '{"source":"manual"}'
  iai dataset-items update item-123 --name ""
```

### Options

```
  -h, --help                   help for update
      --json                   Output raw API response as JSON
      --metadata-json string   Metadata as JSON object (replaces the current metadata)
      --name string            Item name ("" removes it)
  -o, --organization string    Organization name that owns the project
  -p, --project string         Project name
      --status string          Item status (ACTIVE|ARCHIVED)
      --yaml                   Output raw API response as YAML
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

