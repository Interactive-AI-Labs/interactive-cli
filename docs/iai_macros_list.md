## iai macros list

List macros in a project

### Synopsis

List macros in a specific project.

Returns all macros with their name, labels, tags, and last update time.

Examples:
  iai macros list
  iai macros list --page 2 --limit 10

```
iai macros list [flags]
```

### Options

```
  -h, --help                  help for list
      --limit int             Number of items per page (default: 50)
  -o, --organization string   Organization name that owns the project
      --page int              Page number for pagination
  -p, --project string        Project name that owns the prompts
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai macros](iai_macros.md)	 - Manage macros

