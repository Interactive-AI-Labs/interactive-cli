## iai mcps update

Replace an mcp's spec

### Synopsis

Full replace — pass the mcp's complete desired spec, same flags as create.
There is no partial-update mechanism for mcps; anything not passed resets to its
default. The type (internal/external) cannot change; delete and recreate instead.

Changing --credential rotates the mcp's Secret and restarts the mcp (if internal)
and every agent currently attached to it, so they pick up the new value.

```
iai mcps update <mcp_name> [flags]
```

### Examples

```
  iai mcps update my-tool --image-name my-mcp-server --image-tag v2 --port 8080
  iai mcps update acme --external-url https://mcp.acme.com/mcp --credential "$NEW_TOKEN"
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
  -h, --help                        help for update
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

