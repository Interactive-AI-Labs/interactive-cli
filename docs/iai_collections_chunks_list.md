## iai collections chunks list

List chunks (keyset-paginated)

```
iai collections chunks list <collection> [flags]
```

### Examples

```
  iai collections chunks list docs -d my-db --limit 20
  iai collections chunks list docs -d my-db --cursor <token>
```

### Options

```
      --cursor string         Opaque cursor from a previous page
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --limit int             Page size (1-1000, default 100)
  -o, --organization string   Organization name
      --prefix string         Only chunks whose id has this prefix
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

