## iai databases

PostgreSQL instances with extension support, including pgvector

### Synopsis

Manage PostgreSQL databases in InteractiveAI projects.

Databases are managed PostgreSQL instances that can also be used as vector
stores. The "vector" extension (pgvector) is installed by default, enabling
vector similarity search for AI/ML workloads such as RAG and embeddings.

Each database automatically creates a secret named <database_name>-app with
connection credentials (host, port, username, password, URI).

### Options

```
  -h, --help   help for databases
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai databases activate](iai_databases_activate.md)	 - Activate a deactivated database in a project
* [iai databases backup](iai_databases_backup.md)	 - Trigger an on-demand backup
* [iai databases backups](iai_databases_backups.md)	 - List backups for a database
* [iai databases create](iai_databases_create.md)	 - Create a database in a project
* [iai databases deactivate](iai_databases_deactivate.md)	 - Deactivate a database in a project
* [iai databases delete](iai_databases_delete.md)	 - Delete a database from a project
* [iai databases describe](iai_databases_describe.md)	 - Describe a database in detail
* [iai databases list](iai_databases_list.md)	 - List databases in a project
* [iai databases log-fields](iai_databases_log-fields.md)	 - List available fields in structured logs
* [iai databases logs](iai_databases_logs.md)	 - Show logs for a database
* [iai databases port-forward](iai_databases_port-forward.md)	 - Forward a local port to a database
* [iai databases restore](iai_databases_restore.md)	 - Restore a new database from a backup
* [iai databases update](iai_databases_update.md)	 - Update a database in a project

