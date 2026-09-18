## iai skills list

List skills in a project

### Synopsis

List Copilot skills in a project.

Returns all Copilot skills with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

The listing also includes the global skills Interactive loads into every
project. A SCOPE column marks each row as "project" or "global" — the same
values the API takes as ?scope=. Global skills are read-only and cannot be
created, updated or deleted here. They are project-wide and unpaginated, so
they are listed once, on the first page of the root listing — not inside
--folder, and not on a later --page (pages are 0-indexed, so --page 0 is the
first), and they are not counted in the totalCount that --json reports for
paging the project's own skills. A project skill of the same name hides the
global one, matching what the Copilot loads at runtime; a folder of that name
does not, since it renders as "name/".

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

