## iai collections search batch

Run several searches in one request (from a file)

```
iai collections search batch <collection> [flags]
```

### Examples

```
  iai collections search batch docs -d my-db --file searches.json
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --file string           Path to a YAML/JSON batch-search file
  -h, --help                  help for batch
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

* [iai collections search](iai_collections_search.md)	 - Search a collection (single-lane vector search)

