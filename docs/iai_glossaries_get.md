## iai glossaries get

Describe a glossary in detail

### Synopsis

Show detailed information about a specific glossary definition, including its full content.

Without flags, returns the version the server resolves by default. Use
--version to retrieve a specific version number, or --label to resolve a
specific label.

```
iai glossaries get <name> [flags]
```

### Examples

```
  iai glossaries get finance-terms
  iai glossaries get finance-terms --version 3
  iai glossaries get finance-terms --label staging
```

### Options

```
  -h, --help                  help for get
      --json                  Output response as JSON
      --label string          Retrieve the version with this label
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Retrieve a specific version number
      --yaml                  Output response as YAML
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

