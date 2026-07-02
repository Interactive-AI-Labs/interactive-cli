## iai collections chunks patch

Update a chunk's metadata and/or text from a file

```
iai collections chunks patch <collection> <id> [flags]
```

### Examples

```
  iai collections chunks patch docs chunk-1 -d my-db --file patch.json
```

### Options

```
  -d, --database string       Database that holds the collection (required)
      --file string           Path to a YAML/JSON patch file
  -h, --help                  help for patch
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

* [iai collections chunks](iai_collections_chunks.md)	 - Manage the chunks (rows) in a collection

