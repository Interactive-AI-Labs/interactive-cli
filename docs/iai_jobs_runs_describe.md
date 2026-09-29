## iai jobs runs describe

Describe a job run in detail

### Synopsis

Show the latest execution state and explanation for one retained run.

```
iai jobs runs describe <run_id> [flags]
```

### Examples

```
  iai jobs runs describe <run_id>
  iai jobs runs describe <run_id> --watch
  iai jobs runs describe <run_id> --json
```

### Options

```
  -h, --help    help for describe
      --json    Output raw API response as JSON
  -w, --watch   Poll and refresh every 2s until interrupted
      --yaml    Output raw API response as YAML
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

