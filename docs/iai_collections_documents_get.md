## iai collections documents get

Get a document's chunks

```
iai collections documents get <collection> <documentId> [flags]
```

### Examples

```
  iai collections documents get docs support-faq -d my-db
```

### Options

```
      --cursor string         Opaque cursor from a previous page
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for get
      --include-vector        Include the stored vector(s)
      --json                  Output raw API response as JSON
      --limit int             Page size (1-1000, default 100)
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

* [iai collections documents](iai_collections_documents.md)	 - Inspect documents (chunks grouped by documentId)

