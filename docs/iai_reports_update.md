## iai reports update

Replace a report with a document

### Synopsis

Replace a report with a YAML or JSON document, as 'iai reports get --yaml' prints it.

Refused if the report was saved after the document's updated_at; --force overwrites it.
A document with draft_revision, as 'iai reports get' prints your open draft, is saved into
that draft; refused if the draft changed after it was printed, or was saved or discarded.
folder names the report's folder (null for none; a folder that does not exist is created),
and leaving it out keeps the report where it is. Leaving out description, terms, filters
or agent_ids clears them, and leaving out default_range sets last7Days.

```
iai reports update <name> [flags]
```

### Examples

```
  iai reports get "Traffic summary" --yaml > report.yaml
  iai reports update "Traffic summary" --file report.yaml --dry-run
  iai reports update "Traffic summary" --file report.yaml
```

### Options

```
      --dry-run       Print what would change against the saved report; save nothing
      --file string   Path to the YAML or JSON report document; - reads stdin
      --force         Overwrite the report even if it changed after the document's updated_at
  -h, --help          help for update
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

