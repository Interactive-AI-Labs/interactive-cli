## iai mcps update

Update an mcp's spec

### Synopsis

Partial update — only the fields whose flags you pass are changed; everything
else keeps its current value. The type (internal/external) and, for external
mcps, the endpoint/catalog cannot change — delete and recreate instead.

Internal workload flags can be updated independently, except --image-name and
--image-tag, which must be passed together. Changing authentication restarts an
internal mcp and every attached agent. Detach agents before changing auth.

```
iai mcps update <mcp_name> [flags]
```

### Examples

```
  iai mcps update my-tool --image-name my-mcp --image-tag v2
  iai mcps update my-tool --memory 1G --cpu 500m
  iai mcps update acme --auth-type bearer --credential "$NEW_TOKEN"
  iai mcps update acme --description "notes for the team"
```

### Options

```
      --auth-header string          Header used to send the credential
      --auth-header-prefix string   Credential value prefix
      --auth-type string            How the credential is sent: "bearer", "api_key", "none", or "oauth" (inferred on create; required when changing authentication)
      --cpu string                  CPU request/limit, e.g. 250m (internal)
      --credential string           Credential the mcp server requires (bearer token, API key)
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
  -h, --help                        help for update
      --image-name string           Container image name (internal)
      --image-tag string            Container image tag (internal)
      --memory string               Memory request/limit, e.g. 512M (internal)
      --path string                 Endpoint path the mcp's own server exposes (internal, default "/mcp")
      --port int                    Port the mcp server listens on (internal)
      --stack-id string             Stack ID to assign the mcp to (internal)
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

