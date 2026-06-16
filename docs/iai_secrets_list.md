## iai secrets list

List secrets in a project

### Synopsis

List secrets in a specific project.

```
iai secrets list [flags]
```

### Examples

```
  iai secrets list
  iai secrets list -p my-project -o my-org
  iai secrets list --json
```

### Options

```
  -h, --help                  help for list
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

