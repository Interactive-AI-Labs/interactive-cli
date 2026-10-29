## iai agents events create

Send an event to a conversation session

### Synopsis

Send an event to a conversation session on a deployed agent.

```
iai agents events create <agent_name> <session_id> [flags]
```

### Examples

```
  iai agents events create my-agent 9mtKfiDIaZ --message "Where is my order?"
  iai agents events create my-agent 9mtKfiDIaZ --file event.json
  cat event.json | iai agents events create my-agent 9mtKfiDIaZ --file -
```

### Options

```
      --file string           Event as JSON, from a file or stdin (-): a message, custom or status event using the platform API fields, including participant.displayName
  -h, --help                  help for create
      --json                  Output raw API response as JSON
      --message string        Customer message to send; the agent answers it
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output raw API response as YAML
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

