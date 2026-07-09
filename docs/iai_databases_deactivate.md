## iai databases deactivate

Deactivate a database in a project

### Synopsis

Deactivate a database by hibernating it. The database configuration is
preserved and will be restored when the database is activated again.

```
iai databases deactivate <database_name> [flags]
```

### Examples

```
  iai databases deactivate my-db
  iai databases deactivate my-db -p my-project
```

### Options

```
  -h, --help                  help for deactivate
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

