## iai integrations create-custom

Connect a custom MCP endpoint

### Synopsis

Create an integration connection to a custom (user-defined) MCP endpoint.

The connection is verified against the live server on save; if the server cannot
be reached or rejects the credential, creation fails and nothing is stored.

Examples:
  iai integrations create-custom my-server \
    --endpoint-url https://mcp.example.com/mcp --auth-type none
  iai integrations create-custom github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential "$GITHUB_TOKEN"
  iai integrations create-custom github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential-stdin < token.txt
  iai integrations create-custom internal \
    --endpoint-url https://mcp.internal/sse --transport sse \
    --auth-type api_key --credential "$KEY" --header "X-Team=platform"

```
iai integrations create-custom <name> [flags]
```

### Options

```
      --auth-type string      Auth type: api_key, bearer, or none (required)
      --credential string     API key or bearer token (required unless auth-type=none)
      --credential-stdin      Read the credential from stdin instead of --credential
      --description string    Human-readable description
      --endpoint-url string   MCP server endpoint URL (required)
      --header stringArray    Extra header as KEY=VALUE (repeatable)
  -h, --help                  help for create-custom
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connection
      --slug string           Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)
      --transport string      Transport: streamable_http (default) or sse (default "streamable_http")
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

