## iai mcps create

Create an mcp in a project

### Synopsis

Create an mcp — a hosted MCP server ("internal"), a custom external URL,
or a catalog-backed provider.

Internal: --image-name and --image-tag identify the image. --port, --path,
--memory, and --cpu configure how it runs.
External custom: --external-url — a server not owned by the platform, dialed
directly at that URL, path included.
External catalog: --catalog-id (see 'iai mcps catalog'); external URL and auth are
derived from the catalog entry, which provides its own credential header and
prefix. The entry decides the auth type — omit --auth-type unless it accepts
more than one, in which case the error names the options.

The mcp is verified against the live server before it's kept: an internal mcp
is verified automatically once ready; an external mcp (custom or catalog) is verified immediately,
and the create fails if the server is unreachable. Verification lists the
server's tools, so it only catches a bad credential on providers that require
auth to list them — some serve tool discovery anonymously.
An --auth-type oauth mcp is the exception: there is no credential until the
user signs in, so it is created unverified and reports no tools until then.

```
iai mcps create <mcp_name> [flags]
```

### Examples

```
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m --path /api/mcp
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
  iai mcps create github --catalog-id github --credential "$GITHUB_TOKEN"
  iai mcps create github --catalog-id github --credential-stdin < token.txt
  iai mcps create notion --catalog-id notion
  iai mcps create newrelic --catalog-id newrelic --auth-type oauth
```

### Options

```
      --auth-header string          Header used to send the credential
      --auth-header-prefix string   Credential value prefix
      --auth-type string            How the credential is sent: "bearer", "api_key", "none", or "oauth" (inferred on create; required when changing authentication)
      --catalog-id string           Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog external mcp)
      --cpu string                  CPU request/limit, e.g. 250m (internal)
      --credential string           Credential the mcp server requires (bearer token, API key)
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --external-url string         External MCP server URL — not platform-owned, dialed directly (custom external mcp)
  -h, --help                        help for create
      --image-name string           Container image name (internal)
      --image-tag string            Container image tag (internal)
      --memory string               Memory request/limit, e.g. 512M (internal)
      --path string                 Endpoint path the mcp's own server exposes (internal, default "/mcp")
      --port int                    Port the mcp server listens on (internal)
      --stack-id string             Stack ID to assign the mcp to (internal)
      --type string                 Mcp type: "internal" or "external" (inferred from other flags if omitted)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the mcps
```

### SEE ALSO

* [iai mcps](iai_mcps.md)	 - Deploy and manage MCP servers

