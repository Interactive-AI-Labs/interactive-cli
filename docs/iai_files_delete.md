## iai files delete

Delete a file, or one of its versions

### Synopsis

Delete a file and all its versions, or one superseded version with --version.

Deleting a whole file asks for confirmation unless -f is given. Deleting a
single version never asks, since the target is already specific.

```
iai files delete <id|name> [flags]
```

### Examples

```
  iai files delete <id|name>
  iai files delete <id|name> -f
  iai files delete <id|name> --version <version-id>
```

### Options

```
  -f, --force                 Skip the confirmation prompt
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --version string        Delete this specific version instead of the whole file
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

