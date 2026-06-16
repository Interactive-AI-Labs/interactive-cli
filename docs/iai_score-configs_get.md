## iai score-configs get

Get a score config

### Synopsis

Get detailed information about a score configuration.

```
iai score-configs get <id> [flags]
```

### Examples

```
  iai score-configs get sc_123
  iai score-configs get sc_123 -o my-org -p my-project
  iai score-configs get sc_123 --yaml
```

### Options

```
  -h, --help                  help for get
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name that owns the project
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

* [iai score-configs](iai_score-configs.md)	 - Define scoring schemas for evaluation

