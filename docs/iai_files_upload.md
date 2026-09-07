## iai files upload

Upload a document into a project

### Synopsis

Upload a local file into a project as a new document.

The stored name defaults to the local file's base name; use --name to store
it under a different name.

```
iai files upload <local-file> [flags]
```

### Examples

```
  iai files upload ./report.pdf
  iai files upload ./report.pdf --name "Q3 Report.pdf"
  iai files upload ./big.bin --timeout 30m
```

### Options

```
  -h, --help                  help for upload
      --name string           Name to store the file under (default: the local file's name)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name
      --timeout duration      HTTP timeout for the upload (default 15m0s)
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

