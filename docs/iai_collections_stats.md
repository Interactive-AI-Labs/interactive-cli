## iai collections stats

Show a collection's chunk count, size, and index status

```
iai collections stats <collection> [flags]
```

### Examples

```
  iai collections stats docs -d my-db
  iai collections stats docs -d my-db --json
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for stats
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

* [iai collections](iai_collections.md)	 - Knowledge bases (searchable tables of chunks) inside a pgvector database

