## iai services versions

List versions of a service

### Synopsis

Show past versions of a service, sorted newest-first.
Up to 50 versions are retained per service.

Examples:
  iai services versions my-service

```
iai services versions <service_name> [flags]
```

### Options

```
  -h, --help                  help for versions
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

