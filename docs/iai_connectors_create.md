## iai connectors create

Create a connector

### Synopsis

Register an MCP server as a connector, verified against the live server on save.
If the server cannot be reached or rejects the credential, creation fails and
nothing is stored.

Pass --catalog-id to connect a catalog entry (the endpoint and transport come from
the catalog; see 'iai connectors catalog'). Otherwise the connector is custom and
--endpoint-url is required.

Examples:
  iai connectors create github \
    --catalog-id github --auth-type bearer --credential "$GITHUB_TOKEN"
  iai connectors create my-server \
    --endpoint-url https://mcp.example.com/mcp --auth-type none
  iai connectors create github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential-stdin < token.txt
  iai connectors create internal \
    --endpoint-url https://mcp.internal/sse --transport sse \
    --auth-type api_key --credential "$KEY" --header "X-Team=platform"

```
iai connectors create <connector_name> [flags]
```

### Options

```
      --auth-type string      Auth type: api_key, bearer, or none (required)
      --catalog-id string     Catalog entry id for a catalog connector (see 'iai connectors catalog')
      --credential string     API key or bearer token (required unless auth-type=none)
      --credential-stdin      Read the credential from stdin instead of --credential
      --description string    Human-readable description
      --endpoint-url string   MCP server endpoint URL (required for a custom connector)
      --header stringArray    Extra header as KEY=VALUE for a custom connector (repeatable)
  -h, --help                  help for create
      --slug string           Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)
      --transport string      Transport for a custom connector: streamable_http (default) or sse (default "streamable_http")
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the connectors
```

### SEE ALSO

* [iai connectors](iai_connectors.md)	 - Manage MCP connectors in a project

