## iai vector-stores get

Get status of a vector store

### Synopsis

Get the status of a specific vector store.

The project is selected with --project or via 'iai projects select'.

```
iai vector-stores get <vectorStoreName> [flags]
```

### Options

```
  -h, --help                  help for get
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the vector stores
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai vector-stores](iai_vector-stores.md)	 - Manage vector stores

