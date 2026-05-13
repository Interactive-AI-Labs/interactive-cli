## iai glossaries describe

Describe a glossary in detail

### Synopsis

Show detailed information about a specific glossary definition, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai glossaries describe finance-terms
  iai glossaries describe finance-terms --version 3
  iai glossaries describe finance-terms --label staging

```
iai glossaries describe <name> [flags]
```

### Options

```
  -h, --help                  help for describe
      --label string          Retrieve the version with this label (default: server resolves 'production')
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Retrieve a specific version number
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

