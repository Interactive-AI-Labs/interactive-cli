## iai jobs runs list

List runs for a job

### Synopsis

List retained runs of a job in a project, sorted newest-first.

Logs have a separate retention period.

```
iai jobs runs list <job_name> [flags]
```

### Examples

```
  iai jobs runs list my-job
  iai jobs runs list my-job --watch
  iai jobs runs list my-job --json
```

### Options

```
  -h, --help    help for list
      --json    Output raw API response as JSON
  -w, --watch   Poll and refresh the list every 2s until interrupted
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

