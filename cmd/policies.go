package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "policy",
		Plural:   "policies",
		Aliases:  []string{"policy"},
		Short:    "Single-step behavioral rules for agents",
		Long: `Manage policies in InteractiveAI projects.

Policies are core behavior rules — condition-action pairs that govern agent
responses (YAML format).`,
		RouteSegment: "policies",
		HasSchema:    true,
		CreateLong: `Create a new policy in an InteractiveAI project.

Each prompt holds exactly one policy (flat YAML, fields at the root).
Content is provided via a YAML file using the --file flag.
Run 'iai policies schema' to see the current field definitions.

Use --schema-version to validate against a specific schema version. This should
match the schema version of the agent that will use this policy (run
'iai agents compatibility-matrix' to find it). Defaults to the latest stable
schema version when omitted.

Example (policy.yaml):
  id: escalate
  name: Escalation Policy
  condition: User requests human agent
  action: Transfer to human
  criticality: HIGH

The server automatically assigns the "latest" label to new versions. Use
--labels to assign additional labels (e.g. --labels staging).`,
		CreateExample: `  iai policies create safety-rules --file policy.yaml
  iai policies create safety-rules --file policy.yaml --schema-version 2.1.0
  iai policies create safety-rules --file policy.yaml --labels staging
  iai policies create safety-rules --file policy.yaml --tags compliance`,
		ListLong: `List policies in a specific project.

Returns all policies with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.`,
		ListExample: `  iai policies list
  iai policies list --folder my-folder
  iai policies list --folder my-folder/sub-folder
  iai policies list --page 2 --limit 10`,
		GetLong: `Show detailed information about a specific policy, including its full content.

Without flags, returns the version the server resolves by default. Use
--version to retrieve a specific version number, or --label to resolve a
specific label.`,
		GetExample: `  iai policies get safety-rules
  iai policies get safety-rules --version 3
  iai policies get safety-rules --label staging`,
		UpdateLong: `Update a policy by creating a new version with updated content.

This creates a new version of the policy using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

Each prompt holds exactly one policy (flat YAML, fields at the root).
Run 'iai policies schema' to see the current field definitions.

Use --schema-version to validate against a specific schema version. This should
match the schema version of the agent that will use this policy (run
'iai agents compatibility-matrix' to find it). Defaults to the latest stable
schema version when omitted.

Example (policy.yaml):
  id: escalate
  name: Escalation Policy
  condition: User requests human agent
  action: Transfer to human
  criticality: HIGH`,
		UpdateExample: `  iai policies update safety-rules --file policy.yaml
  iai policies update safety-rules --file policy.yaml --schema-version 2.1.0
  iai policies update safety-rules --file policy.yaml --labels staging,qa`,
		DeleteLong: `Delete a policy and all its versions, or delete specific versions.

Without flags, deletes the policy and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.`,
		DeleteExample: `  iai policies delete safety-rules
  iai policies delete safety-rules -f
  iai policies delete safety-rules --version 3
  iai policies delete safety-rules --label staging`,
	})
}
