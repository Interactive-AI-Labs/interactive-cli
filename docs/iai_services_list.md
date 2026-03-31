## iai services list

List services in a project

### Synopsis

List services in a specific project using the deployment service.

```
iai services list [flags]
```

### Options

```
  -h, --help                  help for list
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name to list services from
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

* [iai services](iai_services.md)	 - Manage services

