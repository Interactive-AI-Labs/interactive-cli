## iai databases

PostgreSQL instances with extension support, including pgvector

### Synopsis

Manage PostgreSQL databases in InteractiveAI projects.

Databases are managed PostgreSQL instances that can also be used as vector
stores. The "vector" extension (pgvector) is installed by default, enabling
vector similarity search for AI/ML workloads such as RAG and embeddings.

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
* [iai databases backup](iai_databases_backup.md)	 - Trigger an on-demand backup
* [iai databases backups](iai_databases_backups.md)	 - List backups for a database
* [iai databases create](iai_databases_create.md)	 - Create a database in a project
* [iai databases delete](iai_databases_delete.md)	 - Delete a database from a project
* [iai databases describe](iai_databases_describe.md)	 - Describe a database in detail
* [iai databases list](iai_databases_list.md)	 - List databases in a project
* [iai databases logs](iai_databases_logs.md)	 - Show logs for a database
* [iai databases restore](iai_databases_restore.md)	 - Restore a new database from a backup
* [iai databases update](iai_databases_update.md)	 - Update a database in a project

