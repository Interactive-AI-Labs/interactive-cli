## iai comments list

List comments

### Synopsis

List comments with optional filters.

```
iai comments list [flags]
```

### Options

```
      --author-user-id string   Filter by author user ID
      --columns strings         Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml
  -h, --help                    help for list
      --json                    Output raw API response as JSON
      --limit int               Items per page (max 100)
      --object-id string        Filter by object ID
      --object-type string      Filter by object type (TRACE/OBSERVATION/SESSION/PROMPT)
  -o, --organization string     Organization name that owns the project
      --page int                Page number (starts at 1) (default 1)
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

