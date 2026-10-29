## iai agents sessions list

List conversation sessions on an agent

### Synopsis

List conversation sessions on a deployed agent.

```
iai agents sessions list <agent_name> [flags]
```

### Examples

```
  iai agents sessions list my-agent
  iai agents sessions list my-agent --customer-id cust_42
  iai agents sessions list my-agent --labels test --limit 10
  iai agents sessions list my-agent --json
```

### Options

```
      --cursor string         Cursor from the previous page
      --customer-id string    Only sessions of this customer
  -h, --help                  help for list
      --json                  Output raw API response as JSON
      --labels stringArray    Only sessions carrying this label; repeat the flag once per label
      --limit int             Sessions per page (1-100)
  -o, --organization string   Organization name
  -p, --project string        Project name
      --sort string           Creation order: asc or desc (default asc)
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

