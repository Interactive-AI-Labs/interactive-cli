## iai agents labels set

Set labels on a agent version

### Synopsis

Set labels on an existing version of a agent.

Labels are unique per prompt — assigning a label to this version removes it from
any other version that currently has it.

Examples:
  iai agents labels set my-agent --version 3 --labels production
  iai agents labels set my-agent --version 1 --labels staging,canary

```
iai agents labels set <name> [flags]
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

* [iai agents labels](iai_agents_labels.md)	 - Manage labels for agent versions

