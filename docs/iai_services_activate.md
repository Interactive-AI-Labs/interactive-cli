## iai services activate

Activate a deactivated service in a project

### Synopsis

Activate a deactivated service and restore its previous scale configuration.

```
iai services activate <service_name> [flags]
```

### Examples

```
  iai services activate my-svc
  iai services activate my-svc --project my-project
```

### Options

```
  -h, --help                  help for activate
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

