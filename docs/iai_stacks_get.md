## iai stacks get

Export live stack configuration as a stack YAML file

### Synopsis

Fetch the live services, agents, and databases for a stack and write
them as a stack configuration YAML file.

Use this to rebase your local stack config on the live state before making
changes. The organization and project are read from flags or resolved via
'iai organizations select' / 'iai projects select'.

```
iai stacks get [flags]
```

### Examples

```
  iai stacks get --stack-id my-stack
  iai stacks get --stack-id my-stack -f live-stack.yaml
  iai stacks get --stack-id my-stack -o my-org -p my-project
```

### Options

```
  -f, --file string           Write output to file instead of stdout
  -h, --help                  help for get
      --json                  Output as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --stack-id string       Stack ID to export
      --yaml                  Output as YAML
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

