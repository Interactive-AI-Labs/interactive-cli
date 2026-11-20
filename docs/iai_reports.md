## iai reports

Build and change project reports

### Synopsis

Read and change a project's reports. A report is a page of charts, tables and numbers
over the project's traces, usage and costs, laid out in sections of widgets.

A report is saved whole, as one document: print it with 'iai reports get <name> --yaml',
edit it, and save it with 'iai reports update <name> --file', or 'iai reports create --file'
for a new one. Changing a report requires iai login or JWT authentication.

### Options

```
  -h, --help                  help for reports
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the reports
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai reports create](iai_reports_create.md)	 - Create a report from a document
* [iai reports delete](iai_reports_delete.md)	 - Delete a report
* [iai reports get](iai_reports_get.md)	 - Show a report, or print it as a document to save
* [iai reports list](iai_reports_list.md)	 - List a project's reports
* [iai reports restore](iai_reports_restore.md)	 - Restore a report's previous saved version
* [iai reports schema](iai_reports_schema.md)	 - Show the legal widget selections and the report limits
* [iai reports update](iai_reports_update.md)	 - Replace a report with a document

