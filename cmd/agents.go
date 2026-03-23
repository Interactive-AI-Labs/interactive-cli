package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "agent",
		Plural:   "agents",
		Aliases:  []string{"agent"},
		Short:    "Manage agents",
		Long: `Manage agents in InteractiveAI projects.

Agents are YAML-defined complete agent configurations that bundle identity,
behavioral rules, preamble settings, KB retrieval instructions, resource
references (MCP servers, variables, glossaries, policies), embedded routine
definitions, and priority ordering.`,
		RouteSegment: "agents",
		HasSchema:    true,
		CreateLong: `Create a new agent in an InteractiveAI project.

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
  iai agents create support-agent --file agent.yaml --tags v1,experimental`,
		ListLong: `List agents in a specific project.

Returns all agents with their name, labels, tags, and last update time.

Examples:
  iai agents list
  iai agents list --page 2 --limit 10`,
		GetLong: `Get details of a specific agent, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai agents get support-agent
  iai agents get support-agent --version 3
  iai agents get support-agent --label staging`,
		UpdateLong: `Update an agent by creating a new version with updated content.

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
  iai agents update support-agent --file agent.yaml --labels production,staging`,
		DeleteLong: `Delete an agent and all its versions, or delete specific versions.

Without flags, deletes the agent and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

Examples:
  iai agents delete support-agent
  iai agents delete support-agent -f
  iai agents delete support-agent --version 3
  iai agents delete support-agent --label staging`,
	})
}
