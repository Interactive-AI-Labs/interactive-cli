## iai collections slots progress

Show a slot's index build progress

```
iai collections slots progress <collection> <slot> [flags]
```

### Examples

```
  iai collections slots progress docs title -d my-db
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for progress
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

* [iai collections slots](iai_collections_slots.md)	 - Manage a collection's vector slots and their indexes

