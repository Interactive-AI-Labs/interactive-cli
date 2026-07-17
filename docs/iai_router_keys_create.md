## iai router keys create

Create a router API key

### Synopsis

Create a router API key.

Router keys authenticate inference requests to the InteractiveAI Router, for example chat completions and model calls. They are used as bearer tokens for runtime inference, not for managing project context or infrastructure.

```
iai router keys create <name> [flags]
```

### Options

```
      --description string    Router key description
      --expires-at string     Expiration timestamp (RFC3339). If omitted, keys do not expire by default.
  -h, --help                  help for create
      --json                  Output response as JSON
      --limit float           Credit limit in USD. If omitted, defaults to $100. Maximum is $2500.
      --limit-reset string    Limit reset period: daily, weekly, monthly. If omitted, defaults to monthly.
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai router keys](iai_router_keys.md)	 - Router API keys

