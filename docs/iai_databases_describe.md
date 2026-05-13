## iai databases describe

Describe a database in detail

### Synopsis

Show detailed information about a database including configuration, runtime
status, and connection credentials.

Examples:
  iai databases describe my-db

```
iai databases describe <database_name> [flags]
```

### Options

```
  -h, --help                  help for describe
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

