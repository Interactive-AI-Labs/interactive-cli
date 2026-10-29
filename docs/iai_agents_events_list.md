## iai agents events list

List the events of a conversation session

### Synopsis

List the events of a conversation session on a deployed agent, in offset order.

```
iai agents events list <agent_name> <session_id> [flags]
```

### Examples

```
  iai agents events list my-agent 9mtKfiDIaZ
  iai agents events list my-agent 9mtKfiDIaZ --kinds message
  iai agents events list my-agent 9mtKfiDIaZ --min-offset 12 --follow
  iai agents events list my-agent 9mtKfiDIaZ --follow --json
```

### Options

```
  -f, --follow                Stream new events, updating a reply in place while the agent writes it, for up to 10 minutes; reconnect to continue; cannot combine with --max-offset or --limit
  -h, --help                  help for list
      --json                  Output raw API response as JSON; with --follow, one record per complete event
      --kinds string          Only these kinds, comma-separated: message, tool, status, custom
      --limit int             Return only this many of the newest matching events (1-500); not with --follow
      --max-offset int        Last offset to return; not with --follow
      --min-offset int        First offset to return
  -o, --organization string   Organization name
  -p, --project string        Project name
      --source string         Only events from this source: customer, customerUi, human, humanOnBehalfOfAgent, agent or system
      --trace-id string       Only events of this trace
      --yaml                  Output raw API response as YAML; with --follow, one document per complete event
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

