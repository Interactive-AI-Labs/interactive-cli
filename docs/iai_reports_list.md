## iai reports list

List a project's reports

### Synopsis

List a project's reports with their folder, agents and last change.

```
iai reports list [flags]
```

### Examples

```
  iai reports list
  iai reports list --folder Operations
  iai reports list --agent <agent-id> --json
```

### Options

```
      --agent string      Only reports attached to this agent id
      --columns strings   Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml. Available: id, name, description, folder, agents, default_range, updated_at, updated_by, updated_via, created_at
      --folder string     Only reports in this folder; none for reports in no folder
  -h, --help              help for list
      --json              Output raw API response as JSON
      --yaml              Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the reports
```

### SEE ALSO

* [iai reports](iai_reports.md)	 - Build and change project reports

