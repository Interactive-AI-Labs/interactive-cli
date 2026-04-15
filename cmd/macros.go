package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "macro",
		Plural:   "macros",
		Aliases:  []string{"macro"},
		Short:    "Manage macros",
		Long: `Manage macros in InteractiveAI projects.

Macros are reusable template replies and text snippets (Markdown/text format).
No schema validation is applied — any text content is accepted.`,
		RouteSegment: "macros",
		CreateLong: `Create a new macro in an InteractiveAI project.

Content is provided via a text or markdown file using the --file flag.
No schema validation is applied — any text content is accepted.

Example (disclaimer.md):
  **Disclaimer:** This is not financial advice. Consult a professional.

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai macros create disclaimer --file disclaimer.md
  iai macros create disclaimer --file disclaimer.md --labels production
  iai macros create disclaimer --file disclaimer.md --tags legal`,
		ListLong: `List macros in a specific project.

Returns all macros with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

Examples:
  iai macros list
  iai macros list --folder my-folder
  iai macros list --folder my-folder/sub-folder
  iai macros list --page 2 --limit 10`,
		GetLong: `Get details of a specific macro, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai macros get disclaimer
  iai macros get disclaimer --version 3
  iai macros get disclaimer --label staging`,
		UpdateLong: `Update a macro by creating a new version with updated content.

This creates a new version of the macro using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

No schema validation is applied — any text content is accepted.

Example (disclaimer.md):
  **Disclaimer:** This is not financial advice. Consult a professional.

Examples:
  iai macros update disclaimer --file disclaimer.md
  iai macros update disclaimer --file disclaimer.md --labels production,staging`,
		DeleteLong: `Delete a macro and all its versions, or delete specific versions.

Without flags, deletes the macro and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

Examples:
  iai macros delete disclaimer
  iai macros delete disclaimer -f
  iai macros delete disclaimer --version 3
  iai macros delete disclaimer --label staging`,
	})
}
