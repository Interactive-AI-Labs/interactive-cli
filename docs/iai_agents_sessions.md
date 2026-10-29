## iai agents sessions

Conversation sessions on an agent

### Synopsis

Manage conversation sessions on a deployed agent.

A session holds one conversation. Create one, send it messages with
'iai agents events create' and read the agent's replies with
'iai agents events list', or talk to it from the terminal with 'iai agents chat'.

### Options

```
  -h, --help   help for sessions
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
* [iai agents sessions create](iai_agents_sessions_create.md)	 - Create a conversation session on an agent
* [iai agents sessions list](iai_agents_sessions_list.md)	 - List conversation sessions on an agent

