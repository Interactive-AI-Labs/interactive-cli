## iai mcps connect

Sign in to an mcp that authenticates with your account

### Synopsis

Open a browser and approve access, so the mcp can be used.

Your access token is kept and renewed for you, so this is normally needed once.
Run it again if access is revoked at the provider, or to approve different
permissions.

The account you sign in with is shared — every agent and everyone in the project
uses this connection, the provider's audit log shows your name, and the
connection stops working if your access does.

```
iai mcps connect <mcp_name> [flags]
```

### Examples

```
  iai mcps connect notion-demo
  iai mcps connect notion-demo --no-browser
```

### Options

```
  -h, --help         help for connect
      --no-browser   Print the sign-in URL instead of opening it
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

