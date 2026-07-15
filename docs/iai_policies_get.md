## iai policies get

Describe a policy in detail

### Synopsis

Show detailed information about a specific policy, including its full content.

Without flags, returns the version the server resolves by default. Use
--version to retrieve a specific version number, or --label to resolve a
specific label.

```
iai policies get <name> [flags]
```

### Examples

```
  iai policies get safety-rules
  iai policies get safety-rules --version 3
  iai policies get safety-rules --label staging
```

### Options

```
  -h, --help                  help for get
      --json                  Output response as JSON
      --label string          Retrieve the version with this label
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Retrieve a specific version number
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

* [iai policies](iai_policies.md)	 - Single-step behavioral rules for agents

