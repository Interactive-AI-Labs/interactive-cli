## iai agents update

Update a agent (creates a new version)

### Synopsis

Update an agent by creating a new version with updated content.

This creates a new version of the agent using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

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
  routines:
    - faq-lookup:
        title: FAQ Lookup
        description: Search knowledge base for answers
        steps:
          - id: search
            tools: kb:search
            tool_instruction: Search the FAQ

Examples:
  iai agents update support-agent --file agent.yaml
  iai agents update support-agent --file agent.yaml --labels production,staging

```
iai agents update <name> [flags]
```

### Options

```
      --file string           Path to the file containing the updated prompt content
  -h, --help                  help for update
      --labels strings        Labels for the new prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --tags strings          Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional resource definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Manage agents

