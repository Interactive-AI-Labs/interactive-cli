## iai routines delete

Delete a routine

### Synopsis

Delete a routine and all its versions, or delete specific versions.

Without flags, deletes the routine and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai routines delete onboarding-flow
  iai routines delete onboarding-flow -f
  iai routines delete onboarding-flow --version 3
  iai routines delete onboarding-flow --label staging

```
iai routines delete <name> [flags]
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

* [iai routines](iai_routines.md)	 - Manage routines

