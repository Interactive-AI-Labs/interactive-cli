## iai router-keys update

Update a router API key

```
iai router-keys update <id> [flags]
```

### Options

```
      --clear-limit           Remove the credit limit
      --disable               Disable this key
      --enable                Enable this key
  -h, --help                  help for update
      --json                  Output response as JSON
      --limit float           Credit limit in USD
      --limit-reset string    Limit reset period: none, daily, weekly, monthly
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

* [iai router-keys](iai_router-keys.md)	 - Router API keys

