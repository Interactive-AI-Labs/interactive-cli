## iai jobs runs stop

Stop an unfinished job run

### Synopsis

Request a permanent stop of one unfinished run without disabling future runs.
Returns without waiting for termination; an already-finished run is unchanged.

```
iai jobs runs stop <run_id> [flags]
```

### Examples

```
  iai jobs runs stop <run_id>
  iai jobs runs stop <run_id> --json
```

### Options

```
  -h, --help   help for stop
      --json   Output raw API response as JSON
      --yaml   Output raw API response as YAML
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

