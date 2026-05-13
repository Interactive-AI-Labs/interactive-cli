## iai databases backup

Trigger an on-demand backup

### Synopsis

Trigger an on-demand backup for a database. The database must have backups
enabled.

```
iai databases backup <database_name> [flags]
```

### Options

```
  -h, --help                  help for backup
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

* [iai databases](iai_databases.md)	 - Manage databases

