package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "routine",
		Plural:   "routines",
		Aliases:  []string{"routine"},
		Short:    "Manage routines",
		Long: `Manage routines in InteractiveAI projects.

Routines are step-by-step conversation flows with branching logic and terminal
states (YAML format).`,
		RouteSegment: "routines",
		HasSchema:    true,
		CreateLong: `Create a new routine in an InteractiveAI project.

Content is provided via a YAML file using the --file flag.
Run 'iai routines schema' to see the current field definitions.

Example (routine.yaml):
  title: My Routine
  conditions: When user needs help
  description: Handles user support requests
  steps:
    - id: greet
      description: Welcome the user
      chat_state: Say hello
    - id: lookup
      source: greet
      tools: crm:get_user
      tool_instruction: Fetch user data

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai routines create onboarding-flow --file routine.yaml
  iai routines create onboarding-flow --file routine.yaml --labels production
  iai routines create onboarding-flow --file routine.yaml --tags v2,experimental`,
		ListLong: `List routines in a specific project.

Returns all routines with their name, labels, tags, and last update time.

Examples:
  iai routines list
  iai routines list --page 2 --limit 10`,
		GetLong: `Get details of a specific routine, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai routines get onboarding-flow
  iai routines get onboarding-flow --version 3
  iai routines get onboarding-flow --label staging`,
		UpdateLong: `Update a routine by creating a new version with updated content.

This creates a new version of the routine using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

Run 'iai routines schema' to see the current field definitions.

Example (routine.yaml):
  title: My Routine
  conditions: When user needs help
  description: Handles user support requests
  steps:
    - id: greet
      description: Welcome the user
      chat_state: Say hello
    - id: lookup
      source: greet
      tools: crm:get_user
      tool_instruction: Fetch user data

Examples:
  iai routines update onboarding-flow --file routine.yaml
  iai routines update onboarding-flow --file routine.yaml --labels production,staging`,
		DeleteLong: `Delete a routine and all its versions, or delete specific versions.

Without flags, deletes the routine and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

Examples:
  iai routines delete onboarding-flow
  iai routines delete onboarding-flow -f
  iai routines delete onboarding-flow --version 3
  iai routines delete onboarding-flow --label staging`,
	})
}
