## iai jobs delete

Delete a job from a project

### Synopsis

Delete a job from a specific project using the deployment service.

A job can only be deleted when all its runs have finished.

```
iai jobs delete <job_name> [flags]
```

### Examples

```
  iai jobs delete my-job
  iai jobs delete my-job --project my-project
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

* [iai jobs](iai_jobs.md)	 - Manage saved jobs and executions

