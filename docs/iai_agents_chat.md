## iai agents chat

Talk to an agent from the terminal

### Synopsis

Talk to a deployed agent from the terminal.

Each line you enter is sent as a customer message and the agent's reply is shown
as it is written. Ctrl+C or end of input stops the chat; the session stays on
the agent.

```
iai agents chat <agent_name> [flags]
```

### Examples

```
  iai agents chat my-agent
  iai agents chat my-agent --session-id 9mtKfiDIaZ
  iai agents chat my-agent --customer-id cust_42 --title "Order 123"
```

### Options

```
      --customer-id string    Customer the new conversation belongs to; a guest customer is created when omitted
  -h, --help                  help for chat
  -o, --organization string   Organization name
  -p, --project string        Project name
      --session-id string     Resume this conversation session, showing its latest 50 messages first; a new one is created when omitted
      --title string          Title of the new conversation
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

