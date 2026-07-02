## iai collections documents delete

Delete a document (all chunks sharing the documentId)

```
iai collections documents delete <collection> <documentId> [flags]
```

### Examples

```
  iai collections documents delete docs support-faq -d my-db
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for delete
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

* [iai collections documents](iai_collections_documents.md)	 - Inspect documents (chunks grouped by documentId)

