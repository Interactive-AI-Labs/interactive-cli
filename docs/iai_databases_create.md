## iai databases create

Create a database in a project

### Synopsis

Create a managed PostgreSQL database in a project.

The "vector" extension is installed by default. To add other extensions, use
--extensions. Values above 1 for --instances enable high availability.

Changing the PostgreSQL major version after creation causes cluster downtime
during the upgrade.

```
iai databases create <database_name> [flags]
```

### Examples

```
  iai databases create my-db --instances 2 --cpu 1 --memory 2G --storage-size 20G
  iai databases create my-db --instances 1 --cpu 0.5 --memory 1G --storage-size 20G --extensions vector --extensions pg_trgm
  iai databases create my-db --instances 2 --cpu 1 --memory 2G --storage-size 50G --backup-schedule "0 0 2 * * *" --backup-retention 30d
```

### Options

```
      --backup-retention string   How long to retain backups (e.g. 30d, 4w, 6m)
      --backup-schedule string    Backup schedule as a 6-field cron expression (second minute hour day month weekday, e.g. "0 0 2 * * *" for daily at 02:00)
      --cpu string                CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m); max 7 vCPU (7000m)
      --extensions stringArray    PostgreSQL extension to install (can be repeated); replaces the default list, so include "vector" explicitly if needed; defaults to [vector] if omitted
  -h, --help                      help for create
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

