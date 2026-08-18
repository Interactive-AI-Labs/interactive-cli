## iai stacks diff

Show differences between local config and live stack

### Synopsis

Compare a local stack configuration file against the live state of a
stack and show creates, updates, deletes, and field-level changes.

The local file is read from --file or --cfg-file. The live state is fetched
from the deployment API using --stack-id.

Use --json for machine-readable output in CI pipelines.

```
iai stacks diff [flags]
```

### Examples

```
  iai stacks diff --file stack.yaml --stack-id my-stack
  iai stacks diff --file stack.yaml --stack-id my-stack --json
  iai stacks diff --file stack.yaml --stack-id my-stack -o my-org -p my-project
```

### Options

```
  -f, --file string           Path to local stack configuration file
  -h, --help                  help for diff
      --json                  Output diff as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --stack-id string       Stack ID to compare against live
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

