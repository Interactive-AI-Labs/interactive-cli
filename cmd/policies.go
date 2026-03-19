package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "policy",
		Plural:   "policies",
		Aliases:  []string{"policy"},
		Short:    "Manage policies",
		Long: `Manage policies in InteractiveAI projects.

Policies are core behavior rules — condition-action pairs that govern agent
responses (YAML format).`,
		HasSchema:    true,
		RouteSegment: "policies",
		CreateLong: `Create a new policy in an InteractiveAI project.

Content is provided via a YAML file using the --file flag and must follow the
policy schema below. Use --skip-schema to bypass validation.

Schema:
  policies:                       # required, array of policy rules
    - id: <string>                # required, unique identifier
      condition: <string>         # required, when this rule applies
      action: <string>            # required, what the agent should do
      criticality: <HIGH|MEDIUM|LOW>  # optional, default MEDIUM
      description: <string>       # optional
      tools: [<string>, ...]      # optional, tools to use
      prioritize_over: [<id>, ...]  # optional, policy IDs this overrides

Example (policy.yaml):
  policies:
    - id: p1
      condition: user requests account deletion
      action: confirm identity before proceeding
      criticality: HIGH

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai policies create safety-rules --file policy.yaml
  iai policies create safety-rules --file policy.yaml --labels production
  iai policies create safety-rules --file policy.yaml --tags compliance --skip-schema`,
		ListLong: `List policies in a specific project.

Returns all policies with their name, labels, tags, and last update time.

Examples:
  iai policies list
  iai policies list --page 2 --limit 10`,
		GetLong: `Get details of a specific policy, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai policies get safety-rules
  iai policies get safety-rules --version 3
  iai policies get safety-rules --label staging`,
		UpdateLong: `Update a policy by creating a new version with updated content.

This creates a new version of the policy using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

Schema:
  policies:                       # required, array of policy rules
    - id: <string>                # required, unique identifier
      condition: <string>         # required, when this rule applies
      action: <string>            # required, what the agent should do
      criticality: <HIGH|MEDIUM|LOW>  # optional, default MEDIUM
      description: <string>       # optional
      tools: [<string>, ...]      # optional, tools to use
      prioritize_over: [<id>, ...]  # optional, policy IDs this overrides

Example (policy.yaml):
  policies:
    - id: p1
      condition: user requests account deletion
      action: confirm identity before proceeding
      criticality: HIGH

Examples:
  iai policies update safety-rules --file policy.yaml
  iai policies update safety-rules --file policy.yaml --labels production,staging`,
		DeleteLong: `Delete a policy and all its versions, or delete specific versions.

Without flags, deletes the policy and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

Examples:
  iai policies delete safety-rules
  iai policies delete safety-rules -f
  iai policies delete safety-rules --version 3
  iai policies delete safety-rules --label staging`,
	})
}
