## iai api-keys list

List project API keys

```
iai api-keys list [flags]
```

### Options

```
      --columns strings       Columns to display for table output only (comma-separated, default: id,public_key,secret,note,created_at). Cannot be used with --json or --yaml.
                              Available: id,public_key,secret,note,status,expires_at,last_used_at,created_at
  -h, --help                  help for list
      --json                  Output response as JSON
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

* [iai api-keys](iai_api-keys.md)	 - Project API keys

