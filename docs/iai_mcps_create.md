## iai mcps create

Create an mcp in a project

### Synopsis

Create an MCP server:
  Self-hosted: hosted by the platform; requires --image-name and --image-tag.
  Remote: hosted elsewhere; requires --external-url, including the endpoint path.
  Catalog: a predefined remote provider; --catalog-id supplies its endpoint and auth settings.

On create, a credential defaults to bearer authentication unless --auth-type,
the custom header flags, or a catalog entry with a single auth method select
otherwise. For a catalog entry with several methods, pass --auth-type to use
one other than bearer.

Self-hosted servers receive the credential through the MCP_API_KEY environment
variable; don't set that name with --env. Your server must check the credential
on every request, because the platform doesn't. --endpoint exposes the server
publicly. Agents can only attach bearer or no-auth self-hosted MCPs.

Self-hosted MCPs verify automatically when ready. For remote MCPs, check the
connection after creation with 'iai mcps tools <mcp_name>'; for OAuth, run
'iai mcps connect <mcp_name>' first. Except for client_credentials, a
successful create or anonymous tool discovery doesn't prove the credential
works.

OAuth (--auth-type oauth) uses the signed-in user's identity, even with your
own app. For providers without automatic client registration (e.g. GitHub,
Slack, Google Workspace), register an app and supply --client-id and
--client-secret (or --client-secret-stdin). Create reports the redirect URI
to register and rejects missing registration details before sign-in.
Rotate with 'iai mcps update --auth-type oauth --client-id ...'; this clears
the stored token, so sign in again afterwards.

Machine authentication (--auth-type client_credentials):
  - Requires --catalog-id and a registered app's --client-id and
    --client-secret (or --client-secret-stdin); --external-url is unsupported.
  - The reviewed catalog entry supplies the issuer, token endpoint, scopes,
    and secret delivery method (HTTP Basic or form data).
  - No sign-in: all agents share the app identity. Tokens renew automatically.
  - Creation checks the pair with the provider; rejection leaves no MCP.
    Acceptance does not guarantee tool access; check with 'iai mcps tools'.
    Some providers advertise this grant but require a signed-in user.
  - Signed JWT assertions and client certificates are unsupported.
  - Rotate by deleting and recreating the MCP; tokens minted for the old app
    stop working. 'iai mcps connect' and 'disconnect' apply only to user
    sign-in.

```
iai mcps create <mcp_name> [flags]
```

### Examples

```
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m --path /api/mcp --endpoint
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --env ENV=dev --env SILENT_MODE=true --secret platform-dev
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN" --auth-header X-Token --auth-header-prefix "Token "
  iai mcps create github --catalog-id github --credential "$GITHUB_TOKEN"
  iai mcps create github --catalog-id github --credential-stdin < token.txt
  iai mcps create notion --catalog-id notion
  iai mcps create newrelic --catalog-id newrelic --auth-type oauth
  iai mcps create atlas --catalog-id mongodbatlas --client-id "$CLIENT_ID" --client-secret-stdin < secret.txt
  iai mcps create gh --catalog-id github --auth-type oauth --client-id "$CLIENT_ID" --client-secret-stdin < secret.txt
```

### Options

```
      --auth-header string          Header used for custom authentication
      --auth-header-prefix string   Optional credential prefix for custom authentication (e.g. "Token ")
      --auth-type string            Authentication type: "bearer" (Authorization: Bearer), "api_key" (X-API-Key), "custom", "none", "oauth", or "client_credentials"; inferred on create
      --catalog-id string           Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog remote mcp)
      --client-id string            Client ID of an app you registered at the provider; requires --catalog-id (oauth or client_credentials)
      --client-secret string        Client secret of that app; write-only. Rotatable on oauth via update, delete-and-recreate on client_credentials. Prefer --client-secret-stdin
      --client-secret-stdin         Read the client secret from stdin, keeping it out of shell history and the process list
      --cpu string                  CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m) (self-hosted, default 100m)
      --credential string           Credential the mcp server requires; a self-hosted mcp's server reads it from MCP_API_KEY
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --endpoint                    Expose the mcp publicly at <mcp-name>-<project-hash>.interactive.ai (self-hosted)
      --env stringArray             Environment variable (NAME=VALUE); can be repeated (self-hosted)
      --external-url string         Remote MCP server URL — not platform-owned, dialed directly (custom remote mcp)
  -h, --help                        help for create
      --image-name string           Container image name (self-hosted)
      --image-tag string            Container image tag (self-hosted)
      --memory string               Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G) (self-hosted, default 128M)
      --path string                 Endpoint path the mcp's own server exposes (self-hosted, default "/mcp")
      --port int                    Port the mcp's own server listens on (self-hosted, default 3000)
      --secret stringArray          Existing project secret whose keys become environment variables; repeatable (self-hosted)
      --stack-id string             Stack ID to assign the mcp to (self-hosted)
      --type string                 Mcp type: "self-hosted" or "remote" (inferred from other flags if omitted)
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

