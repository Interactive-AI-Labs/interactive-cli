## iai collections slots add

Add a vector slot

### Synopsis

Add a vector slot. Provide a raw vector slot via flags (--type, --dimension,
--distance) or a full slot config via --file (e.g. for an embedding-backed slot
or custom index tuning). --file takes precedence.

```
iai collections slots add <collection> <slot> [flags]
```

### Examples

```
  iai collections slots add docs title -d my-db --dimension 1536
  iai collections slots add docs title -d my-db --file slot.yaml
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --dimension int         Vector dimension (required unless --file is provided)
      --distance string       Distance metric (default: cosine)
      --file string           Path to a YAML/JSON slot config
  -h, --help                  help for add
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --type string           Vector slot type (float32, float16, binary, or sparse; default: float32) (default "float32")
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

