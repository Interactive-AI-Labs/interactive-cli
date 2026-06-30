## iai collections search by-id

Find neighbors of an existing chunk by its stored vector

```
iai collections search by-id <collection> [flags]
```

### Examples

```
  iai collections search by-id docs -d my-db --id chunk-1 --exclude-self
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --exclude-self          Exclude the seed chunk
      --filter string         Metadata filter as a JSON object
  -h, --help                  help for by-id
      --id string             Seed chunk id (required)
      --json                  Output raw API response as JSON
      --limit int             Max results
  -o, --organization string   Organization name
  -p, --project string        Project name
      --using string          Vector slot to search
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

* [iai collections search](iai_collections_search.md)	 - Search a collection (single-lane vector search)

