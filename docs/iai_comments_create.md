## iai comments create

Create a comment

### Synopsis

Create a new comment on a trace, observation, session, or prompt.

This command requires API key authentication.

```
iai comments create [flags]
```

### Options

```
      --author-user-id string   Author user ID
      --content string          Comment content (required)
  -h, --help                    help for create
      --json                    Output raw API response as JSON
      --object-id string        Object ID (required)
      --object-type string      Object type: TRACE, OBSERVATION, SESSION, or PROMPT (required)
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name
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

* [iai comments](iai_comments.md)	 - Manage comments

