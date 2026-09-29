## iai jobs deactivate

Deactivate a job in a project

### Synopsis

Deactivate a job, preventing new runs. Existing runs continue.
The current configuration is preserved and will be restored when the job is activated again.

```
iai jobs deactivate <job_name> [flags]
```

### Examples

```
  iai jobs deactivate my-job
  iai jobs deactivate my-job --project my-project
```

### Options

```
  -h, --help   help for deactivate
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

