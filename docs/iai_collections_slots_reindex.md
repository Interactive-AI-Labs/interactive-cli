## iai collections slots reindex

Rebuild a slot's index (online)

### Synopsis

Rebuild a slot's index. With no --file it rebuilds with the current config;
--file (YAML/JSON) can change index params or quantization.

```
iai collections slots reindex <collection> <slot> [flags]
```

### Examples

```
  iai collections slots reindex docs title -d my-db
  iai collections slots reindex docs title -d my-db --file reindex.yaml
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --file string           Path to a YAML/JSON reindex config
  -h, --help                  help for reindex
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
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

* [iai collections slots](iai_collections_slots.md)	 - Manage a collection's vector slots and their indexes

