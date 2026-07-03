## iai collections schema

Show the file schemas for the --file-based collection commands

### Synopsis

Print the expected shape of every --file body: collection create/patch,
chunks upsert/patch, slots add/reindex, and search batch/hybrid. Use --json or
--yaml for structured output.

```
iai collections schema [flags]
```

### Examples

```
  iai collections schema
  iai collections schema --json
```

### Options

```
  -h, --help   help for schema
      --json   Output the schemas as JSON
      --yaml   Output the schemas as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai collections](iai_collections.md)	 - Knowledge bases (searchable tables of chunks) inside a pgvector database

