## iai secrets get

Get a secret in a project

### Synopsis

Get a secret in a specific project using the deployment service.

```
iai secrets get <secret_name> [flags]
```

### Examples

```
  iai secrets get my-secret
  iai secrets get my-secret -p my-project
  iai secrets get my-secret --json
```

### Options

```
  -h, --help                  help for get
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the secrets
      --yaml                  Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai secrets](iai_secrets.md)	 - Encrypted key-value pairs for services and agents

