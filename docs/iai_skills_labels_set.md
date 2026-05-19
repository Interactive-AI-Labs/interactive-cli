## iai skills labels set

Set labels on a skill version

### Synopsis

Set labels on an existing skill version, identified by name and version number.

Labels are unique per prompt: assigning a label to one version removes it
from any other version that currently has it. No new version is created.

Examples:
  iai skills labels set my-skill --version 3 --labels production
  iai skills labels set my-skill --version 1 --labels staging,canary

```
iai skills labels set <name> [flags]
```

### Options

```
  -h, --help                  help for set
      --labels strings        Labels to assign (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Version number to assign labels to
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai skills labels](iai_skills_labels.md)	 - Manage labels on skills versions

