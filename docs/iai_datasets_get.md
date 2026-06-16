## iai datasets get

Get a dataset by name

### Synopsis

Get detailed information about a specific dataset.

```
iai datasets get <name> [flags]
```

### Examples

```
  iai datasets get my-dataset
  iai datasets get my-dataset -o my-org -p my-project
  iai datasets get my-dataset --yaml
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

* [iai datasets](iai_datasets.md)	 - Create and list evaluation datasets

