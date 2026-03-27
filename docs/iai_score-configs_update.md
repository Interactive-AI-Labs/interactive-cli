## iai score-configs update

Update a score config

### Synopsis

Update an existing scoring configuration.

This command requires API key authentication.

```
iai score-configs update <id> [flags]
```

### Options

```
      --categories string     New categories as JSON array
      --description string    New description
  -h, --help                  help for update
      --is-archived           Set archived status
      --json                  Output raw API response as JSON
      --max-value float       New maximum value
      --min-value float       New minimum value
  -o, --organization string   Organization name
  -p, --project string        Project name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai score-configs](iai_score-configs.md)	 - Manage score configs

