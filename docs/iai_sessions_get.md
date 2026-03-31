## iai sessions get

Get a specific session

### Synopsis

Get detailed information about a specific session.

Uses the platform API with dual authentication (API key or session).

```
iai sessions get <session-id> [flags]
```

### Options

```
      --fields string         Field groups to include (comma-separated) (default "core,traces")
  -h, --help                  help for get
      --json                  Output raw API response as JSON
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

* [iai sessions](iai_sessions.md)	 - Manage sessions

