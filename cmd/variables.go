package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "variable",
		Plural:   "variables",
		Aliases:  []string{"variable", "var", "vars"},
		Short:    "Manage variables",
		Long: `Manage variables in InteractiveAI projects.

Variables are persistent data fields that survive across conversation sessions
(JSON format).`,
		RouteSegment: "variables",
		HasSchema:    true,
		CreateLong: `Create a new variable definition in an InteractiveAI project.

Content is provided via a JSON file using the --file flag.
Run 'iai variables schema' to see the current field definitions.

Example (variables.json):
  {
    "variables": {
      "user_name": {
        "description": "The user's display name",
        "default_value": "Guest"
      }
    }
  }

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai variables create session-vars --file variables.json
  iai variables create session-vars --file variables.json --labels production
  iai variables create session-vars --file variables.json --tags core`,
		ListLong: `List variables in a specific project.

Returns all variables with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

Examples:
  iai variables list
  iai variables list --folder my-folder
  iai variables list --folder my-folder/sub-folder
  iai variables list --page 2 --limit 10`,
		GetLong: `Get details of a specific variable definition, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai variables get session-vars
  iai variables get session-vars --version 3
  iai variables get session-vars --label staging`,
		UpdateLong: `Update a variable definition by creating a new version with updated content.

This creates a new version of the variable using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

Run 'iai variables schema' to see the current field definitions.

Example (variables.json):
  {
    "variables": {
      "user_name": {
        "description": "The user's display name",
        "default_value": "Guest"
      }
    }
  }

Examples:
  iai variables update session-vars --file variables.json
  iai variables update session-vars --file variables.json --labels production,staging`,
		DeleteLong: `Delete a variable definition and all its versions, or delete specific versions.

Without flags, deletes the variable and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

Examples:
  iai variables delete session-vars
  iai variables delete session-vars -f
  iai variables delete session-vars --version 3
  iai variables delete session-vars --label staging`,
	})
}
