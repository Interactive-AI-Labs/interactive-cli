## iai secrets keys update

Update a single key in a secret

### Synopsis

Update a single key/value pair in an existing secret without replacing the entire secret data.

The project is selected with --project or via 'iai projects select'.

Example:
  iai secrets keys update my-secret -d API_KEY=new-api-key-value

```
iai secrets keys update <secret_name> [flags]
```

### Options

```
  -d, --data string           Secret key data in KEY=VALUE form
  -h, --help                  help for update
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the secrets
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai secrets keys](iai_secrets_keys.md)	 - Manage individual keys within a secret

