## iai prompts labels set

Set labels on a prompt version

### Synopsis

Set labels on an existing version of a prompt.

Labels are unique per prompt — assigning a label to this version removes it from
any other version that currently has it.

Examples:
  iai prompts labels set greeting --version 3 --labels production
  iai prompts labels set greeting --version 1 --labels staging,canary

```
iai prompts labels set <name> [flags]
```

### Options

```
  -h, --help                  help for set
      --labels strings        Labels to assign (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Version number to set labels on
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai prompts labels](iai_prompts_labels.md)	 - Manage labels for prompt versions

