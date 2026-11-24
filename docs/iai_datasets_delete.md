## iai datasets delete

Delete a dataset

### Synopsis

Delete a dataset by name, with all its items and runs. Asks for confirmation unless -f is passed. Needs the session login (iai login); API keys cannot delete datasets.

```
iai datasets delete <name> [flags]
```

### Examples

```
  iai datasets delete my-dataset
  iai datasets delete my-dataset -f -o my-org -p my-project
```

### Options

```
  -f, --force                 Skip confirmation prompt
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
```

### SEE ALSO

* [iai datasets](iai_datasets.md)	 - Create, list and delete evaluation datasets

