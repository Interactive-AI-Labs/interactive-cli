## iai secrets delete

Delete a secret in a project

### Synopsis

Delete a secret in a specific project using the deployment service.

The project is selected with --project or via 'iai projects select'.

```
iai secrets delete <secret_name> [flags]
```

### Options

```
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the secrets
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.dev.interactive.ai")
      --hostname string              Hostname for the API (default "https://dev.interactive.ai")
```

### SEE ALSO

* [iai secrets](iai_secrets.md)	 - Manage secrets

