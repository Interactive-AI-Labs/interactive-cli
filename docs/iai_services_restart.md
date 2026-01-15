## iai services restart

Restart a service in a project

### Synopsis

Restart a service in a specific project using the deployment service.

The project is selected with --project or via 'iai projects select'.

```
iai services restart <service_name> [flags]
```

### Options

```
  -h, --help                  help for restart
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name to restart the service in
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

