## iai routines get

Get details of a routine

### Synopsis

Get details of a specific routine, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai routines get onboarding-flow
  iai routines get onboarding-flow --version 3
  iai routines get onboarding-flow --label staging

```
iai routines get <name> [flags]
```

### Options

```
  -h, --help                  help for get
      --label string          Retrieve the version with this label (default: server resolves 'production')
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Retrieve a specific version number
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional resource definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai routines](iai_routines.md)	 - Manage routines

