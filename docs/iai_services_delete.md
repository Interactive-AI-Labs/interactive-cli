## iai services delete

Delete a service from a project

### Synopsis

Delete a service from a specific project using the deployment service.

The project is selected with --project or via 'iai projects select'.

```
iai services delete [service_name] [flags]
```

### Options

```
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name to delete the service from
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.dev.interactive.ai")
      --hostname string              Hostname for the API (default "https://dev.interactive.ai")
```

### SEE ALSO

* [iai services](iai_services.md)	 - Manage services

