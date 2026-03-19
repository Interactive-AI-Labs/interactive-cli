## iai vector-stores create

Create a vector store

### Synopsis

Create a vector store in a specific project.

```
iai vector-stores create <vectorStoreName> [flags]
```

### Options

```
      --auto-resize             Enable automatic storage resizing
      --auto-resize-limit int   Auto-resize limit in GB (0 = unlimited, requires --auto-resize)
      --backups                 Enable automated daily backups with point-in-time recovery
      --cpu int                 CPU cores (2-80, must be even)
      --ha                      Enable high availability with a standby replica in a separate zone for automatic failover
  -h, --help                    help for create
      --memory float            Memory in GB (2-8 per vCPU, 0.25 increments)
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name that owns the vector stores
      --storage-size int        Storage size in GB, numeric value only (min 20)
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

