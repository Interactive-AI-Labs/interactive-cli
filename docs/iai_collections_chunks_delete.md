## iai collections chunks delete

Delete a single chunk by id

```
iai collections chunks delete <collection> <id> [flags]
```

### Examples

```
  iai collections chunks delete docs chunk-1 -d my-db
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for delete
  -o, --organization string   Organization name
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

* [iai collections chunks](iai_collections_chunks.md)	 - Manage the chunks (rows) in a collection

