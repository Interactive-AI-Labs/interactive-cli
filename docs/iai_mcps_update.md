## iai mcps update

Update an mcp's spec

### Synopsis

Partial update — only the fields whose flags you pass are changed; everything
else keeps its current value. port/path/image/memory/cpu/env/secret only apply
to internal mcps. Use --clear-env, --clear-secret, or --clear-headers to remove
those entirely. The type (internal/external) and, for external mcps, the
endpoint/catalog cannot change — delete and recreate instead.

Changing --credential, or switching --auth-type to "none", rotates the mcp's
Secret and restarts the mcp (if internal) and every agent currently attached
to it. Auth routing cannot change while agents are attached — detach them first.

```
iai mcps update <mcp_name> [flags]
```

### Examples

```
  iai mcps update my-tool --image-tag v2
  iai mcps update my-tool --memory 1G --cpu 500m
  iai mcps update acme --credential "$NEW_TOKEN"
  iai mcps update my-tool --clear-headers
```

### Options

```
      --auth-header string          Header the credential is sent in — only valid with --auth-type custom (bearer/api_key/none each imply their own)
      --auth-header-prefix string   Credential value prefix — only valid with --auth-type custom
      --auth-type string            How the credential is sent: "bearer", "api_key", "custom", or "none" (inferred: "custom" if --auth-header/--auth-header-prefix is set, else "bearer" if --credential is set, else "none")
      --clear-env                   Remove all environment variables from the mcp
      --clear-headers               Remove all extra request headers from the mcp
      --clear-secret                Remove all secret references from the mcp
      --cpu string                  CPU request/limit, e.g. 250m (required for internal)
      --credential string           Credential the mcp server requires (bearer token, API key)
      --credential-stdin            Read the credential from stdin instead of --credential
      --env stringArray             Environment variable (NAME=VALUE) for the mcp server; can be repeated (internal)
      --header stringArray          Extra non-secret request header (NAME=VALUE); can be repeated
  -h, --help                        help for update
      --image-name string           Container image name (internal)
      --image-repository string     Image repository (required for external images)
      --image-tag string            Container image tag (internal)
      --image-type string           Image source: "internal" or "external" (internal) (default "internal")
      --memory string               Memory request/limit, e.g. 512M (required for internal)
      --path string                 Endpoint path the mcp's own server exposes (internal, default "/mcp") — set to whatever the mcp owner actually configured, don't assume
      --port int                    Port the mcp server listens on (internal)
      --secret stringArray          Existing secret to load as env vars; can be repeated (internal)
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

