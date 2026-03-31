## iai score-configs create

Create a score config

### Synopsis

Create a new scoring configuration.

This command requires API key authentication.

```
iai score-configs create [flags]
```

### Options

```
      --categories string     Categories as JSON array
      --data-type string      Data type: NUMERIC, CATEGORICAL, or BOOLEAN (required)
      --description string    Config description
  -h, --help                  help for create
      --json                  Output raw API response as JSON
      --max-value float       Maximum value
      --min-value float       Minimum value
      --name string           Config name (required)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
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

* [iai score-configs](iai_score-configs.md)	 - Manage score configs

