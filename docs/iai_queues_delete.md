## iai queues delete

Delete an annotation queue

### Synopsis

Delete an annotation queue and every item in it.

This command requires Cookie or Bearer authentication; API keys are rejected by the API.

```
iai queues delete <id> [flags]
```

### Examples

```
  iai queues delete queue-123
  iai queues delete queue-123 -o my-org -p my-project
```

### Options

```
  -h, --help                  help for delete
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

* [iai queues](iai_queues.md)	 - Annotation queues for human review workflows

