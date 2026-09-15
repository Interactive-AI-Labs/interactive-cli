## iai mcps port-forward

Forward a local port to an mcp

### Synopsis

Open a local TCP listener and tunnel traffic to an mcp running in
the platform.

The remote port defaults to the mcp's configured port. Use --port to
override. Use --local-port to choose the local listening port (defaults to
--port when set, or an available OS-assigned port otherwise).

```
iai mcps port-forward <mcp_name> [flags]
```

### Examples

```
  iai mcps port-forward my-tool
  iai mcps port-forward my-tool --port 8080
  iai mcps port-forward my-tool --port 8080 --local-port 9090
```

### Options

```
  -h, --help             help for port-forward
      --local-port int   Local port to listen on (defaults to the remote port)
      --port int         Remote port on the mcp (defaults to the mcp's configured port)
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

