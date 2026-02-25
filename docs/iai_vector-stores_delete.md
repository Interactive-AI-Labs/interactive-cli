## iai vector-stores delete

Delete a vector store

### Synopsis

Delete a vector store in a specific project.

The project is selected with --project or via 'iai projects select'.

```
iai vector-stores delete <vectorStoreName> [flags]
```

### Options

```
  -h, --help                  help for delete
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

