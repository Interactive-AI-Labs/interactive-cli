## iai datasets create

Create a dataset

### Synopsis

Create a new evaluation dataset.

```
iai datasets create <name> [flags]
```

### Options

```
      --description string     Dataset description
  -h, --help                   help for create
      --json                   Output raw API response as JSON
      --metadata-json string   Metadata as JSON object
  -o, --organization string    Organization name that owns the project
  -p, --project string         Project name
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

* [iai datasets](iai_datasets.md)	 - Manage evaluation datasets

