## iai replicas list

List replicas for a service

### Synopsis

List pods backing a service in a specific project.

```
iai replicas list [service_name] [flags]
```

### Options

```
  -h, --help                  help for list
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the service
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai replicas](iai_replicas.md)	 - Manage service replicas

