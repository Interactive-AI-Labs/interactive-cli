## iai reports schema

Show the legal widget selections and the report limits

### Synopsis

Show the widget rules (the sources, their metrics and breakdowns, the chart types and
every legal selection) and the limits a report must keep. A widget that matches no
selection is refused on save.

```
iai reports schema [flags]
```

### Examples

```
  iai reports schema
  iai reports schema --json | jq '.selections[] | select(.view == "llms")'
```

### Options

```
  -h, --help   help for schema
      --json   Output raw API response as JSON
      --yaml   Output raw API response as YAML
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

