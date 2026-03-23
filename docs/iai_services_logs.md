## iai services logs

Show logs for a service

### Synopsis

Show logs for all replicas of a service in a project.

Returns up to 5000 log entries in chronological order. Default lookback is 1h.

The project is selected with --project or via 'iai projects select'.

```
iai services logs <service_name> [flags]
```

### Options

```
  -f, --follow                Follow log output
  -h, --help                  help for logs
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the service
      --since string          Relative duration to look back (e.g. 5m, 1h, 3d); default 1h, max 3d
      --start-time string     Absolute RFC3339 timestamp to start from (e.g. 2026-02-24T10:00:00Z); max 3d ago, mutually exclusive with --since
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional resource definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai services](iai_services.md)	 - Manage services

