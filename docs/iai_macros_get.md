## iai macros get

Get details of a macro

### Synopsis

Get details of a specific macro, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai macros get disclaimer
  iai macros get disclaimer --version 3
  iai macros get disclaimer --label staging

```
iai macros get <name> [flags]
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
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai macros](iai_macros.md)	 - Manage macros

