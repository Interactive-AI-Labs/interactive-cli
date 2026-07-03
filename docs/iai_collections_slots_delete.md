## iai collections slots delete

Delete a vector slot

```
iai collections slots delete <collection> <slot> [flags]
```

### Examples

```
  iai collections slots delete docs title -d my-db
```

### Options

```
  -d, --database string       Database that holds the collection (required)
  -h, --help                  help for delete
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

