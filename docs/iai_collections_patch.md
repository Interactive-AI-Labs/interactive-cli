## iai collections patch

Update a collection's mutable config from a file

### Synopsis

Update a collection's mutable configuration from a YAML or JSON file (--file):
full-text settings and per-slot ef_search_default. Slot type/dimension/distance
and the embedding model are immutable.

```
iai collections patch <collection> [flags]
```

### Examples

```
  iai collections patch docs -d my-db --file patch.yaml
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --file string           Path to a YAML/JSON patch config
  -h, --help                  help for patch
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

* [iai collections](iai_collections.md)	 - Vector collections (knowledge bases) inside a pgvector database

