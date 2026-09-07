## iai files restore

Make an earlier version of a file current again

### Synopsis

Make an earlier version current again, without moving its bytes through this client.

The store copies the version's bytes server-side under a new version id; the
source version stays fetchable under its own id, and the file's name is unchanged.

```
iai files restore <id|name> <version-id> [flags]
```

### Examples

```
  iai files restore <id|name> <version-id>
```

### Options

```
  -h, --help                  help for restore
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --timeout duration      HTTP timeout for the restore (default 15m0s)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai files](iai_files.md)	 - Manage documents stored in a project

