## iai policies list

List policies in a project

### Synopsis

List policies in a specific project.

Returns all policies with their name, labels, tags, and last update time.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai policies list
  iai policies list --page 2 --limit 10

```
iai policies list [flags]
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

* [iai policies](iai_policies.md)	 - Manage policies

