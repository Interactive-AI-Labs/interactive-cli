## iai services diff

Compare two revisions of a service

### Synopsis

Show the differences between two revisions of a service.

```
iai services diff <service_name> <revision_a> <revision_b> [flags]
```

### Examples

```
  iai services diff my-service 1 3
```

### Options

```
  -h, --help                  help for diff
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

