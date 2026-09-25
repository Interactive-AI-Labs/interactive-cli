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

The type (self-hosted/remote) and, for remote mcps, the URL/catalog cannot
change — delete and recreate instead. Self-hosted workload flags can be updated
independently. Self-hosted auth fields can be updated independently; remote
credential changes require --auth-type.

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
      --clear-env                   Remove all environment variables from the mcp (self-hosted)
      --clear-secret                Remove all secret references from the mcp (self-hosted)
      --clear-stack-id              Remove the mcp from its stack (self-hosted)
      --client-id string            Client ID of an app you registered at the provider; requires --catalog-id (oauth or client_credentials)
      --client-secret string        Client secret of that app; write-only. Rotatable on oauth via update, delete-and-recreate on client_credentials. Prefer --client-secret-stdin
      --client-secret-stdin         Read the client secret from stdin, keeping it out of shell history and the process list
      --cpu string                  CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m) (self-hosted)
      --credential string           Credential required by the mcp server
      --credential-stdin            Read the credential from stdin instead of --credential
      --description string          Human-readable description of the mcp
      --endpoint                    Expose the mcp publicly (externally accessible) at <mcp-name>-<project-hash>.interactive.ai (self-hosted)
      --env stringArray             Environment variable (NAME=VALUE); can be repeated (self-hosted)
  -h, --help                        help for update
      --image-name string           Container image name (self-hosted)
      --image-tag string            Container image tag (self-hosted)
      --memory string               Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G) (self-hosted)
      --path string                 Endpoint path the mcp's own server exposes (self-hosted, default "/mcp")
      --port int                    MCP port to expose (self-hosted)
      --secret stringArray          Secrets to be loaded as env vars; can be repeated (self-hosted)
      --stack-id string             Stack ID to assign the mcp to (self-hosted)
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

