## iai collections search hybrid

Run a multi-lane hybrid search (RRF) from a file

### Synopsis

Run a hybrid search from a YAML/JSON file whose body holds a "queries" array
(dense and/or full_text lanes) and an optional "fusion" config.

```
iai collections search hybrid <collection> [flags]
```

### Examples

```
  iai collections search hybrid docs -d my-db --file hybrid.json
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --file string           Path to a YAML/JSON hybrid-search file
  -h, --help                  help for hybrid
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

