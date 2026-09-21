## iai mcps create

Create an mcp in a project

### Synopsis

Create an mcp — a hosted MCP server ("internal"), a custom external URL,
or a catalog-backed provider.

Internal: --image-name and --image-tag identify the image. --port, --path,
--memory, and --cpu configure how it runs. --env NAME=VALUE and --secret
configure the server itself and can each be repeated; --secret takes the name
of a secret that already exists in the project (see 'iai secrets'), which is
loaded whole as environment variables. Secret values are never passed here.
External custom: --external-url — a server not owned by the platform, dialed
directly at that URL, path included.
External catalog: --catalog-id (see 'iai mcps catalog'); external URL and auth are
derived from the catalog entry, which provides its own credential header and
prefix. The entry decides the auth type — omit --auth-type unless it accepts
more than one, in which case the error names the options.

An internal mcp is verified automatically once ready. An external mcp is stored
before the platform contacts the provider; after create, run 'iai mcps tools
<mcp_name>' to verify the endpoint and credential. Tool discovery can be
anonymous, so it only catches a bad credential when the provider protects it.
An --auth-type oauth mcp has no credential until the user signs in; connect it
before running the tools check.

Some providers publish no dynamic client registration — GitHub, Slack, Google
Workspace — so there is no app for the sign-in to run under until you register
one yourself. Pass its --client-id and --client-secret alongside --auth-type
oauth and the sign-in runs under your app; the token still belongs to whoever
signs in. 'iai mcps create' names the redirect URI to register when a provider
needs this, and refuses rather than dead-ending at the provider's error page.
Rotating the pair is 'iai mcps update --auth-type oauth --client-id ...', which
drops the stored token, so sign in again afterwards.

An --auth-type client_credentials mcp has no sign-in. You register an app at the
provider and pass its --client-id and --client-secret; the platform mints and
refreshes tokens from that pair. The token is not tied to a person, so every
agent in the project shares one provider identity.

Unlike the other external types, the pair is checked while you wait: a provider
that refuses it fails the create and leaves nothing behind, so a create that
succeeds means the provider accepted the credential.

What client_credentials supports:
  - Catalog entries only. The issuer, token endpoint and scopes come from the
    reviewed entry, so --client-id cannot be combined with --external-url.
  - Providers that accept a client secret at the token endpoint, sent either as
    HTTP Basic or in the form body. The entry decides which.

What it does not support:
  - Providers that only issue tokens to a signed-in person. Many advertise the
    grant and still refuse a machine token; 'iai mcps tools' is how you find out.
  - Signed JWT assertions or client certificates in place of a secret.
  - Changing the pair in place. Rotating means delete and recreate, because
    tokens minted for the old app stop working.
  - 'iai mcps connect' and 'iai mcps disconnect', which are for sign-in types.

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
      --auth-header string          Custom header used to send the credential
      --auth-header-prefix string   Credential value prefix
      --auth-type string            How the credential is sent: "bearer", "api_key", "custom", "none", "oauth", or "client_credentials"; inferred on create
      --catalog-id string           Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog external mcp)
      --client-id string            Client ID of an app you registered at the provider; requires --catalog-id (oauth or client_credentials)
      --client-secret string        Client secret of that app; write-only. Rotatable on oauth via update, delete-and-recreate on client_credentials. Prefer --client-secret-stdin
      --client-secret-stdin         Read the client secret from stdin, keeping it out of shell history and the process list
      --cpu string                  CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m) (internal)
      --credential string           Credential required by the mcp server
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --endpoint                    Expose the mcp at <mcp-name>-<project-hash>.interactive.ai (internal)
      --env stringArray             Environment variable (NAME=VALUE); can be repeated (internal)
      --external-url string         External MCP server URL — not platform-owned, dialed directly (custom external mcp)
  -h, --help                        help for create
      --image-name string           Container image name (internal)
      --image-tag string            Container image tag (internal)
      --memory string               Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G) (internal)
      --path string                 Endpoint path the mcp's own server exposes (internal, default "/mcp")
      --port int                    MCP port to expose (internal)
      --secret stringArray          Secrets to be loaded as env vars; can be repeated (internal)
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

