## iai databases list

List databases in a project

### Synopsis

List databases in a project.

```
iai databases list [flags]
```

### Examples

```
  iai databases list
  iai databases list -p my-project
  iai databases list --json
```

### Options

```
  -h, --help                  help for list
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
  -w, --watch                 Poll and refresh the list every 2s until interrupted
      --yaml                  Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai databases](iai_databases.md)	 - PostgreSQL instances with extension support, including pgvector

