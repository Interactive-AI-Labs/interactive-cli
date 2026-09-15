## iai mcps update

Update an mcp in a project

### Synopsis

Update an mcp in a specific project.

Only the flags you pass are applied; everything else is left at its current
value.

Lists (--env, --secret) replace the entire current list when provided — pass
every value you want to keep.

Use --clear-env or --clear-secret to remove those configurations entirely.

The type (internal/external) and, for external mcps, the endpoint/catalog cannot
change — delete and recreate instead. Internal workload flags can be updated
independently, except --image-name and --image-tag, which must be passed together.
Changing configuration or authentication restarts an internal mcp; an auth change
also restarts every attached agent. Detach agents before changing auth.

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
  iai mcps update acme --auth-type bearer --credential "$NEW_TOKEN"
  iai mcps update acme --description "notes for the team"
```

### Options

```
      --auth-header string          Header used to send the credential
      --auth-header-prefix string   Credential value prefix
      --auth-type string            How the credential is sent: "bearer", "api_key", "none", or "oauth" (inferred on create; required when changing authentication)
      --clear-env                   Remove all environment variables from the mcp (internal)
      --clear-secret                Remove all secret references from the mcp (internal)
      --cpu string                  CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m) (internal)
      --credential string           Credential the mcp server requires (bearer token, API key)
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --endpoint                    Expose the mcp at <mcp-name>-<project-hash>.interactive.ai (internal)
      --env stringArray             Environment variable (NAME=VALUE); can be repeated (internal)
  -h, --help                        help for update
      --image-name string           Container image name (internal)
      --image-tag string            Container image tag (internal)
      --memory string               Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G) (internal)
      --path string                 Endpoint path the mcp's own server exposes (internal, default "/mcp")
      --port int                    MCP port to expose (internal)
      --secret stringArray          Secrets to be loaded as env vars; can be repeated (internal)
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

