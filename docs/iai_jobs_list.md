## iai jobs list

List jobs in a project

### Synopsis

List jobs in a specific project using the deployment service.

```
iai jobs list [flags]
```

### Examples

```
  iai jobs list
  iai jobs list --project my-project
  iai jobs list --json
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

* [iai jobs](iai_jobs.md)	 - Manage saved jobs and executions

