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
		RouteSegment: "policies",
		HasSchema:    true,
		CreateLong: `Create a new policy in an InteractiveAI project.

Each prompt holds exactly one policy (flat YAML, fields at the root).
Content is provided via a YAML file using the --file flag.
Run 'iai policies schema' to see the current field definitions.

Example (policy.yaml):
  id: escalate
  name: Escalation Policy
  condition: User requests human agent
  action: Transfer to human
  criticality: HIGH
  # always_match: true  # evaluate on every turn regardless of context

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai policies create safety-rules --file policy.yaml
  iai policies create safety-rules --file policy.yaml --labels production
  iai policies create safety-rules --file policy.yaml --tags compliance`,
		ListLong: `List policies in a specific project.

Returns all policies with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

Examples:
  iai policies list
  iai policies list --folder my-folder
  iai policies list --folder my-folder/sub-folder
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

Each prompt holds exactly one policy (flat YAML, fields at the root).
Run 'iai policies schema' to see the current field definitions.

Example (policy.yaml):
  id: escalate
  name: Escalation Policy
  condition: User requests human agent
  action: Transfer to human
  criticality: HIGH
  # always_match: true  # evaluate on every turn regardless of context

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
