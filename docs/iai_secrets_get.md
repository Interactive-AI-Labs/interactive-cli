## iai secrets get

Get a secret in a project

### Synopsis

Get a secret in a specific project using the deployment service.

```
iai secrets get <secret_name> [flags]
```

### Options

```
  -h, --help                  help for get
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the secrets
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai secrets](iai_secrets.md)	 - Manage secrets

