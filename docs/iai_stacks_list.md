## iai stacks list

List stacks in a project

### Synopsis

List stacks and their resource counts (services, agents, databases)
in a project. Stacks are discovered from the live resources that belong
to them.

The organization and project are read from flags or resolved via
'iai organizations select' / 'iai projects select'.

```
iai stacks list [flags]
```

### Examples

```
  iai stacks list
  iai stacks list --json
  iai stacks list -o my-org -p my-project
```

### Options

```
  -h, --help                  help for list
      --json                  Output as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai stacks](iai_stacks.md)	 - Declarative resource sync from config files

