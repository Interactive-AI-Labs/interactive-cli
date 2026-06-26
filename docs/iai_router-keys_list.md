## iai router-keys list

List router API keys

```
iai router-keys list [flags]
```

### Options

```
      --columns strings       Columns to display for table output only (comma-separated, default: id,name,key,limit,created_at). Cannot be used with --json or --yaml.
                              Available: id,name,description,status,key,disabled,limit,remaining,limit_reset,expires_at,last_used_at,created_at,updated_at,project_id,user_id
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

* [iai router-keys](iai_router-keys.md)	 - Router API keys

