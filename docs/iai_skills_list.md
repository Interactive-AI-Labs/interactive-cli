## iai skills list

List skills in a project

### Synopsis

List Copilot skills in a project.

Returns all Copilot skills with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

Every skill has a scope, shown in the SCOPE column. Project skills are yours to
edit. Global skills are owned by Interactive and loaded into every project; they
are read-only here. When both scopes have a skill of the same name, the Copilot
uses the project one, and so does this listing.

Global skills are project-wide, so they appear only on the first page of the
root listing, never under --folder, and are not counted in totalCount.

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

