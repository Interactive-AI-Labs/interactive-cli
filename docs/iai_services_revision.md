## iai services revision

Describe a specific revision of a service

### Synopsis

Show the configuration of a specific past revision of a service.

Examples:
  iai services revision my-service 1
  iai services revision my-service 3

```
iai services revision <service_name> <revision> [flags]
```

### Options

```
  -h, --help                  help for revision
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

