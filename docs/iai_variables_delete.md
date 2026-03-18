## iai variables delete

Delete a variable

### Synopsis

Delete a variable definition and all its versions, or delete specific versions.

Without flags, deletes the variable and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai variables delete session-vars
  iai variables delete session-vars -f
  iai variables delete session-vars --version 3
  iai variables delete session-vars --label staging

```
iai variables delete <name> [flags]
```

### Options

```
  -f, --force                 Skip confirmation prompt
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

* [iai variables](iai_variables.md)	 - Manage variables

