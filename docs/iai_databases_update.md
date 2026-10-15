## iai databases update

Update a database in a project

### Synopsis

Partial update of a database. Only the flags you pass are applied; everything
else keeps its current value.

Storage can only be increased. Use --clear-backup to disable backups entirely.
Changing the PostgreSQL major version triggers an automatic upgrade with cluster
downtime.

Use --clear-stack-id to remove the database from its stack.

```
iai databases update <database_name> [flags]
```

### Examples

```
  iai databases update my-db --instances 3
  iai databases update my-db --cpu 2 --memory 4G
  iai databases update my-db --storage-size 50G
  iai databases update my-db --backup-schedule "0 0 3 * * *" --backup-retention 60d
  iai databases update my-db --clear-backup
  iai databases update my-db --stack-id my-stack
  iai databases update my-db --clear-stack-id
```

### Options

```
      --backup-retention string   How long to retain backups (e.g. 30d, 4w, 6m)
      --backup-schedule string    Backup schedule as a 6-field cron expression (second minute hour day month weekday, e.g. "0 0 2 * * *" for daily at 02:00)
      --clear-backup              Remove backup configuration from the database
      --clear-stack-id            Remove the database from its stack
      --cpu string                CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m); max 7 vCPU (7000m)
      --extensions stringArray    PostgreSQL extension to install (can be repeated); replaces the default list, so include "vector" explicitly if needed; defaults to [vector] if omitted
  -h, --help                      help for update
      --instances int             Number of PostgreSQL instances (minimum 1); values above 1 enable high availability
      --memory string             Memory in megabytes (M) or gigabytes (G) (e.g. 512M, 1G, 2G); max 15G
  -o, --organization string       Organization name
      --postgres-version string   PostgreSQL major or major.minor version (e.g. 18, 17.6); supported range 15–18; defaults to latest if omitted
  -p, --project string            Project name
      --stack-id string           Stack ID to assign the database to
      --storage-size string       Storage size with G unit (e.g. 20G, 100G); must be between 10G and 200G; cannot be decreased
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

