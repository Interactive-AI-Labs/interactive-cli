## iai jobs create

Create a job in a project

### Synopsis

Save an image or Python-script job without starting a run.
Use 'iai jobs run' for manual execution.

```
iai jobs create <job_name> [flags]
```

### Examples

```
  iai jobs create my-job --image-type external --image-repository docker.io --image-name python --image-tag 3.12-slim --command python --args=-c --args="print('hello')" --memory 512M --cpu 0.5
  iai jobs create my-job --image-type internal --image-name my-app --image-tag v1 --memory 1G --cpu 1 --env LOG_LEVEL=debug --secret DB_PASSWORD
  iai jobs create report --type script --script main.py --pyproject pyproject.toml --memory 512M --cpu 0.5
  iai jobs create daily-report --type script --script main.py --pyproject pyproject.toml --memory 512M --cpu 0.5 --schedule "0 2 * * *" --schedule-timezone Europe/Berlin
```

### Options

```
      --args stringArray                  Container arguments; can be repeated; image jobs only
      --command stringArray               Container entrypoint override; can be repeated; image jobs only
      --cpu string                        CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)
      --env stringArray                   Environment variable (NAME=VALUE); can be repeated
  -h, --help                              help for create
      --image-name string                 Container image name
      --image-repository string           Image repository; required for external and platform images
      --image-tag string                  Container image tag
      --image-type string                 Image type: 'internal' (project's private registry), 'external' (any public registry), or 'platform' (Interactive AI registries)
      --memory string                     Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)
      --pyproject string                  pyproject.toml path; required for script jobs; combined with --script, maximum 350,000 bytes
      --retention-failed-runs int32       Failed, timed-out, or stopped runs to retain (0-20); server default: 10
      --retention-successful-runs int32   Successful runs to retain (0-20); server default: 5
      --retention-ttl int32               Seconds to keep finished runs (0-2592000); zero requests immediate cleanup; server default: 604800 seconds
      --retries int32                     Maximum retries after failure; share the run timeout and may repeat side effects; server default: 0
      --schedule string                   Five-field cron schedule or calendar shortcut (e.g. '0 2 * * *', '@daily'); omit for manual execution only
      --schedule-timezone string          IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); defaults to UTC; requires --schedule
      --script string                     Python script path; required for script jobs; combined with --pyproject, maximum 350,000 bytes
      --secret stringArray                Secrets to be loaded as env vars; can be repeated
      --stack-id string                   Stack ID to assign the job to
      --timeout int                       Run time budget in seconds (1-21600, max 6h), including startup, dependency installation, and all retries; server default: 3600 seconds
      --type string                       Job type: 'image' or 'script' (default "image")
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

