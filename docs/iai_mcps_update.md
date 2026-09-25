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

The type (internal/external) and, for external mcps, the endpoint/catalog cannot
change — delete and recreate instead. Internal workload flags can be updated
independently. Internal auth fields can be updated independently; external
credential changes require --auth-type, and so does adding a credential to an
internal mcp created without one (e.g. --auth-type bearer). Changing the
credential restarts the mcp and every agent attached to it, so an internal
mcp's server reads the new value from MCP_API_KEY.

```
iai mcps update <mcp_name> [flags]
```

### Examples

```
  iai mcps update my-tool --image-tag v2
  iai mcps update my-tool --memory 1G --cpu 500m
  iai mcps update my-tool --endpoint
  iai mcps update my-tool --endpoint=false
  iai mcps update my-tool --env ENV=dev --env SILENT_MODE=true --secret platform-dev --secret services-dev
  iai mcps update my-tool --clear-env
  iai mcps update my-tool --clear-stack-id
  iai mcps update my-tool --credential-stdin < token.txt
  iai mcps update acme --auth-type bearer --credential "$NEW_TOKEN"
  iai mcps update acme --auth-type custom --credential "$NEW_TOKEN" --auth-header X-Token --auth-header-prefix "Token "
  iai mcps update gh --auth-type oauth --client-id "$NEW_CLIENT_ID" --client-secret-stdin < secret.txt
  iai mcps update acme --description "notes for the team"
```

### Options

```
      --auth-header string          Custom header used to send the credential
      --auth-header-prefix string   Credential value prefix
      --auth-type string            How the credential is sent: "bearer", "api_key", "custom", "none", "oauth", or "client_credentials"; inferred on create
      --clear-env                   Remove all environment variables from the mcp (internal)
      --clear-secret                Remove all secret references from the mcp (internal)
      --clear-stack-id              Remove the mcp from its stack (internal)
      --client-id string            Client ID of an app you registered at the provider; requires --catalog-id (oauth or client_credentials)
      --client-secret string        Client secret of that app; write-only. Rotatable on oauth via update, delete-and-recreate on client_credentials. Prefer --client-secret-stdin
      --client-secret-stdin         Read the client secret from stdin, keeping it out of shell history and the process list
      --cpu string                  CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m) (internal, default 100m)
      --credential string           Credential the mcp server requires; an internal mcp's server reads it from MCP_API_KEY
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --endpoint                    Expose the mcp at <mcp-name>-<project-hash>.interactive.ai (internal)
      --env stringArray             Environment variable (NAME=VALUE); can be repeated (internal)
  -h, --help                        help for update
      --image-name string           Container image name (internal)
      --image-tag string            Container image tag (internal)
      --memory string               Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G) (internal, default 128M)
      --path string                 Endpoint path the mcp's own server exposes (internal, default "/mcp")
      --port int                    Port the mcp's own server listens on (internal, default 3000)
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

