## iai agents create

Create a agent

### Synopsis

Create a new agent in an InteractiveAI project.

Content is provided via a YAML file using the --file flag.
Run 'iai agents schema' to see the current field definitions.

Example (agent.yaml):
  agent_id: support-agent
  description: Customer support agent with FAQ and escalation
  extra_rules:
    - Always greet the customer by name
    - Escalate billing issues immediately
  preamble:
    greeting: "Hello! How can I help you today?"
    language_instruction: "Respond in the customer's language"
  mcp_servers:
    - crm-server
  variables:
    - business-hours
  glossaries:
    - product-catalog
  policies:
    - data-privacy
  routines:
    - faq-lookup:
        title: FAQ Lookup
        description: Search knowledge base for answers
        steps:
          - id: search
            tools: kb:search
            tool_instruction: Search the FAQ
    - escalate:
        title: Escalate
        description: Hand off to human agent
        steps:
          - id: handoff
            chat_state: Transfer to human support

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai agents create support-agent --file agent.yaml
  iai agents create support-agent --file agent.yaml --labels production
  iai agents create support-agent --file agent.yaml --tags v1,experimental

```
iai agents create <name> [flags]
```

### Options

```
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --tags strings          Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Manage agents

