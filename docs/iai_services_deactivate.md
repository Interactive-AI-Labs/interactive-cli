## iai services deactivate

Deactivate a service in a project

### Synopsis

Deactivate a service, stopping all running instances. The current configuration
is preserved and will be restored when the service is activated again.

```
iai services deactivate <service_name> [flags]
```

### Examples

```
  iai services deactivate my-svc
  iai services deactivate my-svc --project my-project
```

### Options

```
  -h, --help                  help for deactivate
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

* [iai services](iai_services.md)	 - Deploy and manage HTTP services

