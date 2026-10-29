## iai agents events get

Get one event of a conversation session

### Synopsis

Get one event of a conversation session on a deployed agent.

```
iai agents events get <agent_name> <session_id> <event_id> [flags]
```

### Examples

```
  iai agents events get my-agent 9mtKfiDIaZ 3f9c2a
  iai agents events get my-agent 9mtKfiDIaZ 3f9c2a --follow
  iai agents events get my-agent 9mtKfiDIaZ 3f9c2a --json
```

### Options

```
  -f, --follow                Stream a message the agent is still writing, updated in place until it is complete, for up to 10 minutes; reconnect to continue
  -h, --help                  help for get
      --json                  Output raw API response as JSON; with --follow, the event once it is complete
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output raw API response as YAML; with --follow, the event once it is complete
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents events](iai_agents_events.md)	 - Events in a conversation session on an agent

