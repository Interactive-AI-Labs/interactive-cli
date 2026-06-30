## iai collections chunks bulk-delete

Delete many chunks by ids, metadata filter, or all

### Synopsis

Delete chunks by exactly one selector: --ids, --filter, or --all.

--all deletes every chunk and requires confirmation.

```
iai collections chunks bulk-delete <collection> [flags]
```

### Examples

```
  iai collections chunks bulk-delete docs -d my-db --ids a,b,c
  iai collections chunks bulk-delete docs -d my-db --filter '{"lang":"en"}'
  iai collections chunks bulk-delete docs -d my-db --all
```

### Options

```
      --all                   Delete every chunk (requires confirm)
  -d, --database string       Database that holds the collection (required)
      --filter string         Metadata filter as a JSON object
  -h, --help                  help for bulk-delete
      --ids strings           Comma-separated chunk ids to delete
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output raw API response as YAML
      --yes                   Skip the --all confirmation prompt
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

