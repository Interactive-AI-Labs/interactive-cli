## iai collections create

Create a collection from a config file

### Synopsis

Create a vector collection from a YAML or JSON config file (--file).

The config declares the vector slot(s) — either an embedding-backed slot
("embedding": {model, dimension}) or a raw vector slot ({type, dimension,
distance}) — and optional full-text search.

Slot type, dimension, distance, and the embedding model are IMMUTABLE after
creation; fixing a wrong value means deleting and recreating the collection.

Run 'iai collections schema' for the config file format.

```
iai collections create <collection> [flags]
```

### Examples

```
  iai collections create docs -d my-db --file collection.yaml
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --dry-run               Validate the config without creating the collection
      --file string           Path to a YAML/JSON collection config
  -h, --help                  help for create
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

