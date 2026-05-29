## iai integrations create-from-catalog

Connect an MCP server from the catalog

### Synopsis

Create an integration connection from a curated catalog entry. The endpoint and
transport come from the catalog entry; you supply a name and (unless the entry
needs no auth) a credential.

Use 'iai integrations catalog' to find the --catalog-id.

The connection is verified against the live server on save.

Examples:
  iai integrations create-from-catalog github \
    --catalog-id github --auth-type bearer --credential "$GITHUB_TOKEN"

```
iai integrations create-from-catalog <name> [flags]
```

### Options

```
      --auth-type string      Auth type: api_key, bearer, or none (required)
      --catalog-id string     Catalog entry id (required; see 'iai integrations catalog')
      --credential string     API key or bearer token (required unless auth-type=none)
      --description string    Human-readable description
  -h, --help                  help for create-from-catalog
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connection
      --slug string           Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai integrations](iai_integrations.md)	 - MCP integration connections for a project

