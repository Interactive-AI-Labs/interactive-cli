## iai collections slots vacuum

Vacuum a slot (reclaim space, refresh stats)

```
iai collections slots vacuum <collection> <slot> [flags]
```

### Examples

```
  iai collections slots vacuum docs title -d my-db
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for vacuum
  -o, --organization string   Organization name
  -p, --project string        Project name
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

