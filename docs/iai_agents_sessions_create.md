## iai agents sessions create

Create a conversation session on an agent

### Synopsis

Create a conversation session on a deployed agent and print its ID.

```
iai agents sessions create <agent_name> [flags]
```

### Examples

```
  iai agents sessions create my-agent
  iai agents sessions create my-agent --customer-id cust_42 --title "Order 123"
  iai agents sessions create my-agent --labels test --json
```

### Options

```
      --allow-greeting        Let the agent open the conversation with a greeting
      --customer-id string    Customer the conversation belongs to; a guest customer is created when omitted
  -h, --help                  help for create
      --json                  Output raw API response as JSON
      --labels stringArray    Label attached to the conversation; repeat the flag once per label
  -o, --organization string   Organization name
  -p, --project string        Project name
      --title string          Conversation title
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

* [iai agents sessions](iai_agents_sessions.md)	 - Conversation sessions on an agent

