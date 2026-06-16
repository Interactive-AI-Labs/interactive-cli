## iai replicas describe

Describe a replica in detail

### Synopsis

Show detailed information about a specific replica including status, resources, healthcheck configuration, and events.

```
iai replicas describe <replica_name> [flags]
```

### Examples

```
  iai replicas describe my-service-abc123
  iai replicas describe my-service-abc123 -p my-project -o my-org
  iai replicas describe my-service-abc123 --yaml
```

### Options

```
  -h, --help                  help for describe
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the service
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

* [iai replicas](iai_replicas.md)	 - Inspect service replicas

