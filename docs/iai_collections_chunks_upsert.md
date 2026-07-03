## iai collections chunks upsert

Upsert chunks from a file

### Synopsis

Upsert a batch of chunks from a YAML or JSON file (--file).

Chunks with text and no client vector are embedded server-side (set
defer_embedding=true with client-supplied vectors to skip embedding).

```
iai collections chunks upsert <collection> [flags]
```

### Examples

```
  iai collections chunks upsert docs -d my-db --file chunks.json
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --dry-run               Validate the batch without embedding or storing
      --file string           Path to a YAML/JSON chunks file
  -h, --help                  help for upsert
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

* [iai collections chunks](iai_collections_chunks.md)	 - Manage the chunks (rows) in a collection

