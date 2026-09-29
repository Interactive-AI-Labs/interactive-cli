## iai jobs activate

Activate a deactivated job in a project

### Synopsis

Activate a deactivated job, allowing new runs and restoring its schedule, if configured.

```
iai jobs activate <job_name> [flags]
```

### Examples

```
  iai jobs activate my-job
  iai jobs activate my-job --project my-project
```

### Options

```
  -h, --help   help for activate
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

