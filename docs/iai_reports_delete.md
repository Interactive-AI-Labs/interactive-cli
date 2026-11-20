## iai reports delete

Delete a report

### Synopsis

Delete a report. Prints the deleted report as a document, so 'iai reports create'
can create it again.

```
iai reports delete <name> [flags]
```

### Examples

```
  iai reports delete "Traffic summary"
  iai reports delete "Traffic summary" -f > deleted-report.yaml
```

### Options

```
  -f, --force   Skip the confirmation prompt
  -h, --help    help for delete
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

