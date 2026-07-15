## iai mcps create

Create an mcp in a project

### Synopsis

Create an mcp — an in-cluster MCP server ("internal"), a custom external
URL, or a catalog-backed provider.

Internal: --image-name, --image-tag, --port (like 'iai services create'); --env
and --secret load env vars from literal values or existing k8s Secrets.
External custom: --external-url — a server not owned by the platform, dialed
directly at that URL.
External catalog: --catalog-id (see 'iai mcps catalog'); external URL and auth are
derived from the catalog entry. Pass an auth type the entry supports; catalog
entries provide their own credential header and prefix.

The mcp is verified against the live server before it's kept: an internal mcp
is verified once its pod is up (checked in the background — see 'iai mcps get');
an external mcp (custom or catalog) is verified immediately, and the create
fails if the server is unreachable or rejects the credential.

```
iai mcps create <mcp_name> [flags]
```

### Examples

```
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
  iai mcps create github --catalog-id github --credential "$GITHUB_TOKEN"
  iai mcps create github --catalog-id github --credential-stdin < token.txt
```

### Options

```
      --auth-header string          Header the credential is sent in — only valid with --auth-type custom (bearer/api_key/none each imply their own)
      --auth-header-prefix string   Credential value prefix — only valid with --auth-type custom
      --auth-type string            How the credential is sent: "bearer", "api_key", "custom", or "none" (inferred: "custom" if --auth-header/--auth-header-prefix is set, else "bearer" if --credential is set, else "none")
      --catalog-id string           Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog external mcp)
      --cpu string                  CPU request/limit, e.g. 250m (required for internal)
      --credential string           Credential the mcp server requires (bearer token, API key)
      --credential-stdin            Read the credential from stdin instead of --credential
      --env stringArray             Environment variable (NAME=VALUE) for the mcp server; can be repeated (internal)
      --external-url string         External MCP server URL — not platform-owned, dialed directly (custom external mcp)
      --header stringArray          Extra non-secret request header (NAME=VALUE); can be repeated
  -h, --help                        help for create
      --image-name string           Container image name (internal)
      --image-repository string     Image repository (required for external images)
      --image-tag string            Container image tag (internal)
      --image-type string           Image source: "internal" or "external" (internal) (default "internal")
      --memory string               Memory request/limit, e.g. 512M (required for internal)
      --port int                    Port the mcp server listens on (internal)
      --secret stringArray          Existing k8s Secret to load as env vars; can be repeated (internal)
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

