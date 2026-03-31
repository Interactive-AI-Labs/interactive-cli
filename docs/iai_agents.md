## iai agents

Manage agents

### Synopsis

Manage agents in InteractiveAI projects.

Agents are YAML-defined complete agent configurations that bundle identity,
behavioral rules, preamble settings, KB retrieval instructions, resource
references (MCP servers, variables, glossaries, policies), embedded routine
definitions, and priority ordering.

### Options

```
  -h, --help   help for agents
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai agents create](iai_agents_create.md)	 - Create a agent
* [iai agents delete](iai_agents_delete.md)	 - Delete a agent
* [iai agents get](iai_agents_get.md)	 - Get details of a agent
* [iai agents list](iai_agents_list.md)	 - List agents in a project
* [iai agents schema](iai_agents_schema.md)	 - Display the JSON Schema for agents
* [iai agents update](iai_agents_update.md)	 - Update a agent (creates a new version)

