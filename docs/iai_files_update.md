## iai files update

Upload a new version of a file

### Synopsis

Upload local-file as a new version of an existing document.

The file keeps its stored name unless --name is given.

```
iai files update <id|name> <local-file> [flags]
```

### Examples

```
  iai files update <id|name> ./report.pdf
  iai files update <id|name> ./report.pdf --name "Q3 Report.pdf"
```

### Options

```
  -h, --help                  help for update
      --name string           Rename the file as part of this update (default: keep its stored name)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --timeout duration      HTTP timeout for the update (default 15m0s)
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

