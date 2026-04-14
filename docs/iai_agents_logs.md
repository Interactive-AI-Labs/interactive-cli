## iai agents logs

Show logs for an agent

### Synopsis

Show logs for an agent in a project.

Returns up to 5000 log entries in chronological order.

Examples:
  iai agents logs my-agent
  iai agents logs my-agent --follow
  iai agents logs my-agent --since 30m
  iai agents logs my-agent --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z

```
iai agents logs <agent_name> [flags]
```

### Options

```
      --end-time string       Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow
  -f, --follow                Stream new log entries as they arrive; mutually exclusive with --end-time
  -h, --help                  help for logs
  -o, --organization string   Organization name
  -p, --project string        Project name
      --since string          Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time
      --start-time string     Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Manage agents

