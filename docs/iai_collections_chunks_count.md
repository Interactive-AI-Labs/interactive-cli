## iai collections chunks count

Count chunks, optionally scoped by a metadata filter or id prefix

```
iai collections chunks count <collection> [flags]
```

### Examples

```
  iai collections chunks count docs -d my-db --filter '{"lang":"en"}'
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --filter string         Metadata filter as a JSON object
  -h, --help                  help for count
  -o, --organization string   Organization name
      --prefix string         Only count chunks with this id prefix
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

