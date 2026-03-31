## iai dataset-runs delete

Delete a dataset run

### Synopsis

Delete a dataset run by name.

```
iai dataset-runs delete <run-name> [flags]
```

### Options

```
      --dataset-name string   Dataset name (required)
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai dataset-runs](iai_dataset-runs.md)	 - Manage dataset runs

