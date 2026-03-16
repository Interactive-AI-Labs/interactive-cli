## iai macros delete

Delete a macro

### Synopsis

Delete a macro and all its versions, or delete specific versions.

Without flags, deletes the macro and all its versions. Use --version to delete
a specific version, or --label to delete versions with a specific label.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai macros delete disclaimer
  iai macros delete disclaimer --version 3
  iai macros delete disclaimer --label staging

```
iai macros delete <name> [flags]
```

### Options

```
  -h, --help                  help for delete
      --label string          Delete versions with this label only
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Delete a specific version only
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

