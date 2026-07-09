## iai databases activate

Activate a deactivated database in a project

### Synopsis

Activate a deactivated database.

```
iai databases activate <database_name> [flags]
```

### Examples

```
  iai databases activate my-db
  iai databases activate my-db -p my-project
```

### Options

```
  -h, --help                  help for activate
  -o, --organization string   Organization name
  -p, --project string        Project name
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

