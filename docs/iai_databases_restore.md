## iai databases restore

Restore a new database from a backup

### Synopsis

Create a new database by restoring from an existing database's backup. The
source database must have backups enabled.

Optionally specify --target-time for point-in-time recovery (RFC3339 format).
If omitted, the latest backup is restored.

Examples:
  iai databases restore my-restored-db --source-database my-db --instances 2 --cpu 1 --memory 2G --storage-size 20G
  iai databases restore my-restored-db --source-database my-db --target-time 2026-05-12T10:00:00Z --instances 2 --cpu 1 --memory 2G --storage-size 20G

```
iai databases restore <database_name> [flags]
```

### Options

```
      --backup-retention string   How long to retain backups (e.g. 30d, 4w, 6m)
      --backup-schedule string    Backup schedule as a 6-field cron expression (second minute hour day month weekday)
      --cpu string                CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)
      --extensions stringArray    PostgreSQL extension to install (can be repeated); defaults to [vector] if omitted
  -h, --help                      help for restore
      --instances int             Number of PostgreSQL instances (minimum 1); values above 1 enable high availability
      --memory string             Memory in megabytes (M) or gigabytes (G) (e.g. 512M, 1G, 2G)
  -o, --organization string       Organization name
      --postgres-version string   PostgreSQL major or major.minor version (e.g. 17, 16.4); defaults to latest if omitted
  -p, --project string            Project name
      --source-database string    Name of the database to restore from; must have backups enabled
      --storage-size string       Storage size with G unit (e.g. 20G, 100G); must be between 10G and 200G; cannot be decreased
      --target-time string        RFC3339 timestamp for point-in-time recovery (e.g. 2026-05-12T10:00:00Z); omit to restore the latest backup
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai databases](iai_databases.md)	 - Manage databases

