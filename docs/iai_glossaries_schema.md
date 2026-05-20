## iai glossaries schema

Display the JSON Schema for glossaries

### Synopsis

Fetch and display the JSON Schema for glossaries from the backend API.

Use --schema-version to request a specific schema version (defaults to latest stable).
Use --json to output the raw JSON Schema instead of a formatted table.

This is a public endpoint and does not require authentication.

```
iai glossaries schema [flags]
```

### Options

```
  -h, --help                    help for schema
      --json                    Output raw JSON Schema instead of a formatted table
      --schema-version string   Schema version to fetch (defaults to latest stable)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai glossaries](iai_glossaries.md)	 - Domain vocabularies for consistent term interpretation

