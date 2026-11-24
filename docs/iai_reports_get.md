## iai reports get

Show a report, or print it as a document to save

### Synopsis

Show a report's settings, sections and widgets. In the Copilot, while the report is open
in the user's editor, it shows the editor's unsaved draft instead, and --yaml and --json print
it with its draft_revision.

```
iai reports get <name|id> [flags]
```

### Examples

```
  iai reports get "Traffic summary"
  iai reports get "Traffic summary" --yaml > report.yaml
```

### Options

```
  -h, --help   help for get
      --json   Output the report as a JSON document that 'iai reports update' reads back
      --yaml   Output the report as a YAML document that 'iai reports update' reads back
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

