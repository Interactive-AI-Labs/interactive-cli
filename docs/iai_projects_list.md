## iai projects list

List projects in an organization

### Synopsis

List all projects within a specific organization. The organization name will be resolved to its Id before making API calls.

```
iai projects list [flags]
```

### Examples

```
  iai projects list
  iai projects list --organization my-org
  iai projects list --json
```

### Options

```
  -h, --help                  help for list
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the projects
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

* [iai projects](iai_projects.md)	 - Switch or list projects

