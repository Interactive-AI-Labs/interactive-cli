## iai agents events

Events in a conversation session on an agent

### Synopsis

Read and create the events of a conversation session on a deployed agent.

A conversation is an ordered list of events: the messages each side sends,
status events as the agent works on a reply, tool events and custom events.
Each event has an offset that orders it in the conversation.

### Options

```
  -h, --help   help for events
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools
* [iai agents events create](iai_agents_events_create.md)	 - Send an event to a conversation session
* [iai agents events get](iai_agents_events_get.md)	 - Get one event of a conversation session
* [iai agents events list](iai_agents_events_list.md)	 - List the events of a conversation session

