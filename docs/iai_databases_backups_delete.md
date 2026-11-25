## iai databases backups delete

Delete a database backup

### Synopsis

Delete a backup and its stored data. Other backups are kept and can still be
restored. Backups in progress cannot be deleted.

```
iai databases backups delete <database_name> <backup_name> [flags]
```

### Examples

```
  iai databases backups delete my-db my-db-on-demand-1795614308
  iai databases backups delete my-db my-db-on-demand-1795614308 --force
```

### Options

```
  -f, --force                 Skip confirmation prompt
  -h, --help                  help for delete
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

* [iai databases backups](iai_databases_backups.md)	 - List backups for a database

