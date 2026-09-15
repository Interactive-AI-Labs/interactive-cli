## iai replicas list

List replicas for a service or MCP

### Synopsis

List pods backing a service (default) or MCP (--type mcp) in a project.

```
iai replicas list <resource_name> [flags]
```

### Examples

```
  iai replicas list my-service
  iai replicas list my-service -p my-project -o my-org
  iai replicas list my-service --json
  iai replicas list my-tool --type mcp
```

### Options

```
  -h, --help                  help for list
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the workload
      --yaml                  Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --type string                  Replica workload type: service or mcp (default "service")
```

### SEE ALSO

* [iai replicas](iai_replicas.md)	 - Inspect service or MCP replicas

