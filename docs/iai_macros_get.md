## iai macros get

Describe a macro in detail

### Synopsis

Show detailed information about a specific macro, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

```
iai macros get <name> [flags]
```

### Examples

```
  iai macros get disclaimer
  iai macros get disclaimer --version 3
  iai macros get disclaimer --label staging
```

### Options

```
  -h, --help                  help for get
      --json                  Output response as JSON
      --label string          Retrieve the version with this label (default: server resolves 'production')
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

* [iai macros](iai_macros.md)	 - Pre-approved response templates used in routines

