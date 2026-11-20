## iai reports restore

Restore a report's previous saved version

### Synopsis

Swap a report with the version saved before its last change. Restoring twice brings
the change back.

```
iai reports restore <name> [flags]
```

### Examples

```
  iai reports restore "Traffic summary"
  iai reports restore "Traffic summary" --expected-updated-at 2026-11-10T09:30:00Z
```

### Options

```
      --expected-updated-at string   Fail without restoring unless the report's updated_at equals this value ('iai reports get --yaml' prints it); defaults to the updated_at read just before
  -h, --help                         help for restore
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

