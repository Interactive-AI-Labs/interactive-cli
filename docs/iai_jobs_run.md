## iai jobs run

Run a job in a project

### Synopsis

Start a run using the job's saved configuration.

The job must be active and have no unfinished run. The command returns the run
without waiting for completion; use 'iai jobs runs describe --watch' to follow its status.

```
iai jobs run <job_name> [flags]
```

### Examples

```
  iai jobs run my-job
  iai jobs run my-job --project my-project
  iai jobs run my-job --json
```

### Options

```
  -h, --help   help for run
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

* [iai jobs](iai_jobs.md)	 - Manage saved jobs and executions

