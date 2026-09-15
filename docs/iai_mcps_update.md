## iai mcps update

Update an mcp in a project

### Synopsis

Update an mcp in a specific project.

Only the flags you pass are applied; everything else is left at its current
value.

Lists (--env, --secret) replace the entire current list when provided — pass
every value you want to keep.

Use --clear-env, --clear-secret, or --clear-stack-id to remove those
configurations entirely.

The hosting type (internal: platform-hosted; external: hosted elsewhere) cannot
change. For external mcps, the URL/catalog also cannot change — delete and
recreate instead. Internal workload flags can be updated independently, except
--image-name and --image-tag, which must be passed together.
Internal auth fields can be updated independently; external credential changes
require --auth-type.

```
iai mcps update <mcp_name> [flags]
```

### Examples

```
  iai mcps update my-tool --image-name my-mcp --image-tag v2
  iai mcps update my-tool --memory 1G --cpu 500m
  iai mcps update my-tool --endpoint
  iai mcps update my-tool --endpoint=false
  iai mcps update my-tool --env ENV=dev --env SILENT_MODE=true --secret platform-dev --secret services-dev
  iai mcps update my-tool --clear-env
  iai mcps update my-tool --clear-stack-id
  iai mcps update my-tool --credential-stdin < token.txt
  iai mcps update acme --auth-type bearer --credential "$NEW_TOKEN"
  iai mcps update acme --description "notes for the team"
```

### Options

```
      --auth-header string          Header used to send the credential
      --auth-header-prefix string   Credential value prefix
      --auth-type string            How the credential is sent: "bearer", "api_key", "custom" (internal), "none", or "oauth" (external); inferred on create
      --clear-env                   Remove all environment variables from the mcp (internal only; platform-hosted)
      --clear-secret                Remove all secret references from the mcp (internal only; platform-hosted)
      --clear-stack-id              Remove the mcp from its stack (internal only; platform-hosted)
      --cpu string                  CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m) (internal only; platform-hosted)
      --credential string           Credential the mcp server requires (bearer token, API key)
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --endpoint                    Expose the mcp at <mcp-name>-<project-hash>.interactive.ai (internal only; platform-hosted)
      --env stringArray             Environment variable (NAME=VALUE); can be repeated (internal only; platform-hosted)
  -h, --help                        help for update
      --image-name string           Container image name (internal only; platform-hosted)
      --image-tag string            Container image tag (internal only; platform-hosted)
      --memory string               Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G) (internal only; platform-hosted)
      --path string                 Endpoint path the mcp's own server exposes, default "/mcp" (internal only; platform-hosted)
      --port int                    MCP port to expose (internal only; platform-hosted)
      --secret stringArray          Secrets to be loaded as env vars; can be repeated (internal only; platform-hosted)
      --stack-id string             Stack ID to assign the mcp to (internal only; platform-hosted)
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

