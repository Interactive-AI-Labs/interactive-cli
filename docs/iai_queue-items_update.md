## iai queue-items update

Update a queue item

### Synopsis

Update the status of a queue item.

This command requires API key authentication.

```
iai queue-items update <item-id> [flags]
```

### Examples

```
  iai queue-items update item-456 --queue-id queue-123 --status COMPLETED
  iai queue-items update item-456 --queue-id queue-123 --status PENDING --json
  iai queue-items update item-456 --queue-id queue-123 --status COMPLETED --yaml
```

### Options

```
  -h, --help                  help for update
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --queue-id string       Queue ID (required)
      --status string         New status (required)
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

* [iai queue-items](iai_queue-items.md)	 - Manage items in annotation queues

