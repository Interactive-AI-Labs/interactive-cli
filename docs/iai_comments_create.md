## iai comments create

Create a comment

### Synopsis

Create a new comment on a trace, observation, session, or prompt.

This command requires API key authentication.

```
iai comments create [flags]
```

### Examples

```
  iai comments create --object-type TRACE --object-id trace-abc123 --content "Investigated this run"
  iai comments create --object-type OBSERVATION --object-id obs-456 --content "Looks correct" --author-user-id user-42
  iai comments create --object-type PROMPT --object-id prompt-789 --content "Needs review" --json
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
      --yaml                    Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai comments](iai_comments.md)	 - Annotate traces, observations, and sessions

