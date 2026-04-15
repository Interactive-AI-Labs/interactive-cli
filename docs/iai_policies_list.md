## iai policies list

List policies in a project

### Synopsis

List policies in a specific project.

Returns all policies with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

Examples:
  iai policies list
  iai policies list --folder my-folder
  iai policies list --folder my-folder/sub-folder
  iai policies list --page 2 --limit 10

```
iai policies list [flags]
```

### Options

```
      --folder string         List items inside the given folder path
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

