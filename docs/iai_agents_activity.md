## iai agents activity

List an agent's deploys, context changes and secret edits

### Synopsis

List an agent's activity: deploys, context changes and secret edits, most recent first.

Without agent_name, the selected agent is used (see 'iai agents select').

```
iai agents activity [agent_name] [flags]
```

### Examples

```
  iai agents activity my-agent
  iai agents activity my-agent --search deploy --limit 50
  iai agents activity my-agent --actor user:123 --actor api_key:iai_ab --json
```

### Options

```
      --actor stringArray     Filter by actor: user:<id>, api_key:<prefix> or system:<component> (repeatable; --json lists them under actors)
  -h, --help                  help for activity
      --json                  Output raw API response as JSON
      --limit int             Items per page (max 100)
  -o, --organization string   Organization name
      --page int              Page number (starts at 1) (default 1)
  -p, --project string        Project name
      --search string         Search in actor, action and data (max 200 characters)
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

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools

