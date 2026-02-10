## iai replicas describe

Describe a replica in detail

### Synopsis

Show detailed information about a specific replica including status, resources, healthcheck configuration, and events.

The project is selected with --project or via 'iai projects select'.

```
iai replicas describe <replica_name> [flags]
```

### Options

```
  -h, --help                  help for describe
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the service
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai replicas](iai_replicas.md)	 - Manage service replicas

