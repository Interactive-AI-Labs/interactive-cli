## iai jobs update

Update a job in a project

### Synopsis

Update supplied fields of a saved job; omitted fields keep their values.
Preserves activation state and run history. All runs must be finished.

```
iai jobs update <job_name> [flags]
```

### Examples

```
  iai jobs update my-job --image-tag v2
  iai jobs update my-job --image-tag v2 --expect-revision 3
  iai jobs update my-job --memory 1G --cpu 0.5
  iai jobs update report --script main.py
  iai jobs update report --schedule "0 2 * * *" --schedule-timezone Europe/Berlin
  iai jobs update report --clear-schedule
  iai jobs update report --retention-successful-runs 0
```

### Options

```
      --args stringArray                  Replace image arguments; repeat for every argument to keep; image jobs only
      --clear-args                        Remove the container arguments override
      --clear-command                     Remove the container entrypoint override
      --clear-env                         Remove all environment variables from the job
      --clear-retention                   Reset retention to its defaults
      --clear-schedule                    Remove the schedule configuration from the job
      --clear-secret                      Remove all secret references from the job
      --clear-stack-id                    Remove the job from its stack
      --command stringArray               Replace the entrypoint; repeat for every argument to keep; image jobs only
      --cpu string                        CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)
      --env stringArray                   Replace environment variables (NAME=VALUE); repeat for every entry to keep; dropping entries requires --force
      --expect-revision int               Require this live revision before updating; also fail if the revision cannot be fetched
      --force                             Allow --env/--secret to drop live entries; --clear-env/--clear-secret need no override. Drop checks are skipped if live state cannot be fetched
  -h, --help                              help for update
      --image-name string                 Container image name
      --image-repository string           Image repository; required for external and platform images
      --image-tag string                  Container image tag
      --image-type string                 Image type: 'internal' (project's private registry), 'external' (any public registry), or 'platform' (Interactive AI registries)
      --memory string                     Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)
      --pyproject string                  Replace pyproject.toml from a local file; combined with the new or retained script, maximum 350,000 bytes
      --retention-failed-runs int32       Failed, timed-out, or stopped runs to retain (0-20); unchanged when omitted
      --retention-successful-runs int32   Successful runs to retain (0-20); unchanged when omitted
      --retention-ttl int32               Seconds to keep finished runs (0-2592000); zero requests immediate cleanup; unchanged when omitted
      --retries int32                     Maximum retries after failure; share the run timeout and may repeat side effects; unchanged when omitted
      --schedule string                   Five-field cron schedule or calendar shortcut (e.g. '0 2 * * *', '@daily')
      --schedule-timezone string          IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC)
      --script string                     Replace the script from a local file; combined with the new or retained project file, maximum 350,000 bytes
      --secret stringArray                Replace secret references; repeat for every entry to keep; dropping entries requires --force
      --stack-id string                   Stack ID to assign the job to
      --timeout int                       Run time budget in seconds (1-21600, max 6h), including startup, dependency installation, and all retries; unchanged when omitted
      --type string                       Job type: 'image' or 'script'; clears the old type's fields; supply the new type's inputs
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

