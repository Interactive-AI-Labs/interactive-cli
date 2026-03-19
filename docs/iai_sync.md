## iai sync

Sync services and vector stores from a stack config file

### Synopsis

Sync services and vector stores in a project from a stack configuration file.

For services, sync will:
- Create services that exist in the config but not in the project
- Update services that exist in both the config and the project
- Delete services that exist in the project but not in the config (for the specified stack)

For vector stores, sync will:
- Create vector stores that exist in the config but not in the project
- Delete vector stores that exist in the project but not in the config (for the specified stack)

The project is selected with --project or via 'iai projects select', and the config file with --cfg-file.

```
iai sync [flags]
```

### Options

```
  -h, --help                  help for sync
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name to sync in
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI

