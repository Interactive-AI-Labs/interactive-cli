## iai connectors

Manage MCP connectors in a project

### Synopsis

A connector stores the endpoint, transport, and credentials for an MCP server.
Pick one from the platform catalog or define a custom endpoint. Once connected,
verify to discover available tools and run them directly.

### Options

```
  -h, --help                  help for connectors
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the connectors
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
* [iai connectors catalog](iai_connectors_catalog.md)	 - Browse the connector catalog
* [iai connectors create](iai_connectors_create.md)	 - Create a connector
* [iai connectors delete](iai_connectors_delete.md)	 - Delete a connector
* [iai connectors get](iai_connectors_get.md)	 - Show a connector and its tools
* [iai connectors list](iai_connectors_list.md)	 - List connectors in a project
* [iai connectors run-tool](iai_connectors_run-tool.md)	 - Run a tool on a connector
* [iai connectors verify](iai_connectors_verify.md)	 - Re-verify a connector and refresh its tools

