## iai skills list

List skills in a project

### Synopsis

List Copilot skills in a project.

Returns all Copilot skills with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

```
iai skills list [flags]
```

### Examples

```
  iai skills list
  iai skills list --folder my-folder
  iai skills list --page 2 --limit 10
```

### Options

```
      --folder string         List items inside the given folder path
  -h, --help                  help for list
      --json                  Output response as JSON
      --limit int             Number of items per page (default: 50)
  -o, --organization string   Organization name that owns the project
      --page int              Page number for pagination
  -p, --project string        Project name that owns the prompts
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

* [iai skills](iai_skills.md)	 - Manage Interactive Copilot skills (not to be confused with context items that configure the Interactive Agent)

