## iai queue-items delete

Delete a queue item

### Synopsis

Delete an item from an annotation queue.

```
iai queue-items delete <item-id> [flags]
```

### Options

```
  -h, --help                  help for delete
  -o, --organization string   Organization name
  -p, --project string        Project name
      --queue-id string       Queue ID (required)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai queue-items](iai_queue-items.md)	 - Manage annotation queue items

