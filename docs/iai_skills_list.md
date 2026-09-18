## iai skills list

List skills in a project

### Synopsis

List Copilot skills in a project.

A skill has one of two scopes, shown in the SCOPE column. Project skills are
yours to edit. Global skills are owned by Interactive and served to every
project; they are read-only here. Both scopes are listed, so a name that exists
in both appears twice, once per scope — the Copilot runtime loads the project
one at conversation time.

The listing is complete rather than paginated. Folders are shown with a trailing
"/" (colored when stdout is a terminal) and can be browsed into with --folder,
which lists the project alone: global skills are project-wide and sit in no
folder.

```
iai skills list [flags]
```

### Examples

```
  iai skills list
  iai skills list --folder my-folder
  iai skills list --json
```

### Options

```
      --folder string         List items inside the given folder path
  -h, --help                  help for list
      --json                  Output response as JSON
  -o, --organization string   Organization name that owns the project
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

