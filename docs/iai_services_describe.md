## iai services describe

Describe a service in detail

### Synopsis

Show detailed information about a specific service including its configuration.

Examples:
  iai services describe my-service

```
iai services describe <service_name> [flags]
```

### Options

```
  -h, --help                  help for describe
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

* [iai services](iai_services.md)	 - Manage services

