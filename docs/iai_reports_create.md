## iai reports create

Create a report from a document

### Synopsis

Create a report from a YAML or JSON document, as 'iai reports get --yaml' prints it.

The document's updated_at is ignored, so a printed report saves as a copy under another
name. folder names the report's folder, and a folder that does not exist is created.
Leaving out default_range sets last7Days. Fields only the platform sets, such as
created_at and updated_by, are ignored.

```
iai reports create [flags]
```

### Examples

```
  iai reports get "Traffic summary" --yaml > report.yaml
  iai reports create --file report.yaml --dry-run
  iai reports create --file - < report.yaml
```

### Options

```
      --dry-run       Print the report that would be created; save nothing
      --file string   Path to the YAML or JSON report document; - reads stdin
  -h, --help          help for create
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

