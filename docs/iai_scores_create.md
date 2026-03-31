## iai scores create

Create a score

### Synopsis

Create a score on exactly one target resource.

This command currently requires API key authentication.

```
iai scores create [flags]
```

### Options

```
      --comment string          Score comment
      --config-id string        Related config ID
      --data-type string        Score data type (default "NUMERIC")
      --environment string      Target environment
  -h, --help                    help for create
      --id string               Explicit score ID
      --json                    Output raw API response as JSON
      --metadata-json string    Metadata as JSON object
      --name string             Score name
      --observation-id string   Target observation ID
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name
      --queue-id string         Related queue ID
      --session-id string       Target session ID
      --trace-id string         Target trace ID
      --value string            Score value
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

* [iai scores](iai_scores.md)	 - Manage scores

