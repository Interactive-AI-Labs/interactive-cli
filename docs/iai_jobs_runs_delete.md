## iai jobs runs delete

Delete a finished job run

### Synopsis

Delete a finished run from history. Unfinished runs cannot be deleted.

```
iai jobs runs delete <run_id> [flags]
```

### Examples

```
  iai jobs runs delete <run_id>
  iai jobs runs delete <run_id> --project my-project
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name
```

### SEE ALSO

* [iai jobs runs](iai_jobs_runs.md)	 - Inspect and manage job runs

