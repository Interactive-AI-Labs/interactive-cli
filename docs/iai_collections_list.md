## iai collections list

List collections in a database

```
iai collections list [flags]
```

### Examples

```
  iai collections list -d my-db
  iai collections list -d my-db --json
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for list
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

* [iai collections](iai_collections.md)	 - Vector collections (knowledge bases) inside a pgvector database

