## iai mcps

Deploy and manage MCP servers

### Synopsis

Manage MCP servers for a project — in-cluster workloads ("internal"), custom
external URLs, or catalog-backed providers (external, external URL + auth derived
from the curated catalog).

Attach an mcp to an agent with '--mcp <name>' on 'iai agents create'/'update'.

### Options

```
  -h, --help                  help for mcps
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the mcps
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
* [iai mcps catalog](iai_mcps_catalog.md)	 - Browse the curated MCP catalog
* [iai mcps create](iai_mcps_create.md)	 - Create an mcp in a project
* [iai mcps delete](iai_mcps_delete.md)	 - Delete an mcp
* [iai mcps get](iai_mcps_get.md)	 - Show mcp details, verify state, and cached tools
* [iai mcps list](iai_mcps_list.md)	 - List mcps in a project
* [iai mcps run-tool](iai_mcps_run-tool.md)	 - Run a tool on an mcp
* [iai mcps tools](iai_mcps_tools.md)	 - List an mcp's cached tools with descriptions
* [iai mcps update](iai_mcps_update.md)	 - Replace an mcp's spec
* [iai mcps verify](iai_mcps_verify.md)	 - Re-verify an external mcp and refresh its cached tools

