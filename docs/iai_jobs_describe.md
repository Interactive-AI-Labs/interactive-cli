## iai jobs describe

Describe a job in detail

### Synopsis

Show detailed information about a specific job including its configuration.

For script jobs, the script and project file are summarized; use --script or
--pyproject to print that file exactly as stored, e.g. to save it locally.

```
iai jobs describe <job_name> [flags]
```

### Examples

```
  iai jobs describe my-job
  iai jobs describe my-job --json
  iai jobs describe report --script > main.py
  iai jobs describe report --pyproject > pyproject.toml
```

### Options

```
  -h, --help        help for describe
      --json        Output raw API response as JSON
      --pyproject   Print only the pyproject.toml of a script job
      --script      Print only the script of a script job
  -w, --watch       Poll and refresh every 2s until interrupted
      --yaml        Output raw API response as YAML
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

