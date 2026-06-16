## iai databases port-forward

Forward a local port to a database

### Synopsis

Open a local TCP listener and tunnel traffic through the deployment operator
to a PostgreSQL database running in the cluster.

The remote port defaults to 5432. Use --port to override. Use --local-port
to choose the local listening port (defaults to the remote port).

After connecting you can use psql, pgAdmin, or any PostgreSQL client against
localhost:<local-port>.

```
iai databases port-forward <database_name> [flags]
```

### Examples

```
  iai databases port-forward my-db
  iai databases port-forward my-db --local-port 15432
```

### Options

```
  -h, --help                  help for port-forward
      --local-port int        Local port to listen on (defaults to the remote port)
  -o, --organization string   Organization name
      --port int              Remote port on the database (defaults to 5432)
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

