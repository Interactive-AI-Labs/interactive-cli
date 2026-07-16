## iai models get

Get a router model

### Synopsis

Get detailed information about a router model by its ID.

```
iai models get <id> [flags]
```

### Examples

```
  iai models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411
  iai models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411 -o my-org -p my-project
  iai models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411 --json
  iai models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411 --yaml
```

### Options

```
  -h, --help                  help for get
      --json                  Output response as JSON
  -o, --organization string   Organization name that owns the project
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

* [iai models](iai_models.md)	 - List and inspect models

