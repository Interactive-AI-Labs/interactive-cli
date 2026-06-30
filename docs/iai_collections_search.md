## iai collections search

Search a collection (single-lane vector search)

### Synopsis

Run a single-lane search: --query (text, embedded server-side) or --vector
(comma-separated floats). --exact runs an exhaustive scan instead of the index.

Sub-commands cover the other modes: batch, by-id, hybrid.

```
iai collections search <collection> [flags]
```

### Examples

```
  iai collections search docs -d my-db --query "reset my password"
  iai collections search docs -d my-db --query "..." --exact --limit 5
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --exact                 Exhaustive scan instead of the index
      --filter string         Metadata filter as a JSON object
  -h, --help                  help for search
      --json                  Output raw API response as JSON
      --limit int             Max results
  -o, --organization string   Organization name
  -p, --project string        Project name
      --query string          Query text (embedded server-side)
      --using string          Vector slot to search (default: default)
      --vector string         Query vector as comma-separated floats
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

* [iai collections](iai_collections.md)	 - Vector collections (knowledge bases) inside a pgvector database
* [iai collections search batch](iai_collections_search_batch.md)	 - Run several searches in one request (from a file)
* [iai collections search by-id](iai_collections_search_by-id.md)	 - Find neighbors of an existing chunk by its stored vector
* [iai collections search hybrid](iai_collections_search_hybrid.md)	 - Run a multi-lane hybrid search (RRF) from a file

