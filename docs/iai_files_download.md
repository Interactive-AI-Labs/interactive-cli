## iai files download

Download a file's bytes

### Synopsis

Download a file's current version, or an older one with --version.

Without --output, the file is written under the name the server reports,
reduced to a safe local filename; --output - streams to stdout instead.

```
iai files download <id> [flags]
```

### Examples

```
  iai files download <id>
  iai files download <id> --version <version-id>
  iai files download <id> --output report.pdf
  iai files download <id> --output - > report.pdf
```

### Options

```
  -f, --force                 Overwrite an existing local file
  -h, --help                  help for download
  -o, --organization string   Organization name that owns the project
      --output string         Local path to write to (default: the stored name); - for stdout
  -p, --project string        Project name
      --timeout duration      HTTP timeout for the download (default 15m0s)
      --version string        Download this specific version instead of the current one
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

