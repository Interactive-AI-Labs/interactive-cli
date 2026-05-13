## iai variables describe

Describe a variable in detail

### Synopsis

Show detailed information about a specific variable definition, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai variables describe session-vars
  iai variables describe session-vars --version 3
  iai variables describe session-vars --label staging

```
iai variables describe <name> [flags]
```

### Options

```
  -h, --help                  help for describe
      --label string          Retrieve the version with this label (default: server resolves 'production')
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Retrieve a specific version number
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai variables](iai_variables.md)	 - Contextual attributes referenced in policies and routines

