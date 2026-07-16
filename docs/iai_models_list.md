## iai models list

List router models

### Synopsis

List router models for a project.

```
iai models list [flags]
```

### Examples

```
  iai models list
  iai models list -o my-org -p my-project
  iai models list --page 1 --limit 10
  iai models list --search claude
  iai models list --region eu
  iai models list --json
  iai models list --yaml
```

### Options

```
  -h, --help                  help for list
      --json                  Output response as JSON
      --limit int             Items per page (max 100) (default 50)
  -o, --organization string   Organization name that owns the project
      --page int              Page number (0-indexed)
  -p, --project string        Project name
      --region string         Filter by region (us|eu)
      --search string         Search filter
      --yaml                  Output response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai models](iai_models.md)	 - List and inspect models

