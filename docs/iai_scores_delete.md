## iai scores delete

Delete a score

### Synopsis

Delete a score by ID.

This command currently requires API key authentication.

```
iai scores delete <score-id> [flags]
```

### Options

```
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
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

* [iai scores](iai_scores.md)	 - Manage scores

