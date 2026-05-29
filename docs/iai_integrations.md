## iai integrations

MCP integration connections for a project

### Synopsis

Manage Model Context Protocol (MCP) integration connections in an InteractiveAI project.

Connections let agents reach external tools exposed by an MCP server — either a
curated catalog entry (a vendor-hosted server) or a custom endpoint you define.
Create a connection, verify it to discover its tools, then run a tool directly.

### Options

```
  -h, --help   help for integrations
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai integrations catalog](iai_integrations_catalog.md)	 - Browse the MCP integrations catalog
* [iai integrations create-custom](iai_integrations_create-custom.md)	 - Connect a custom MCP endpoint
* [iai integrations create-from-catalog](iai_integrations_create-from-catalog.md)	 - Connect an MCP server from the catalog
* [iai integrations delete](iai_integrations_delete.md)	 - Delete an integration connection
* [iai integrations get](iai_integrations_get.md)	 - Show an integration connection and its tools
* [iai integrations list](iai_integrations_list.md)	 - List integration connections in a project
* [iai integrations tools](iai_integrations_tools.md)	 - Inspect and run tools on a connection
* [iai integrations verify](iai_integrations_verify.md)	 - Re-verify a connection and refresh its tools

