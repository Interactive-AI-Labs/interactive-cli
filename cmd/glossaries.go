package cmd

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "glossary",
		Plural:   "glossaries",
		Aliases:  []string{"glossary"},
		Short:    "Manage glossary definitions",
		Long: `Manage glossary definitions in InteractiveAI projects.

Glossary entries are domain-specific terms with descriptions and synonyms (JSON
format).`,
		RouteSegment: "glossaries",
		CreateLong: `Create a new glossary definition in an InteractiveAI project.

Content is provided via a JSON file using the --file flag and must follow the
glossary schema below.

Schema:
  {"terms": [                     // required, array of glossary terms
    {
      "name": "<string>",         // required, the term
      "description": "<string>",  // required, definition of the term
      "synonyms": ["<string>"]    // optional, alternative names
    }
  ]}

Example (glossary.json):
  {"terms": [
    {"name": "APR", "description": "Annual Percentage Rate", "synonyms": ["annual rate"]},
    {"name": "KYC", "description": "Know Your Customer"}
  ]}

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai glossaries create finance-terms --file glossary.json
  iai glossaries create finance-terms --file glossary.json --labels production
  iai glossaries create finance-terms --file glossary.json --tags domain`,
		ListLong: `List glossary definitions in a specific project.

Returns all glossary entries with their name, labels, tags, and last update time.

Examples:
  iai glossaries list
  iai glossaries list --page 2 --limit 10`,
		GetLong: `Get details of a specific glossary definition, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai glossaries get finance-terms
  iai glossaries get finance-terms --version 3
  iai glossaries get finance-terms --label staging`,
		UpdateLong: `Update a glossary definition by creating a new version with updated content.

This creates a new version of the glossary using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.

Schema:
  {"terms": [                     // required, array of glossary terms
    {
      "name": "<string>",         // required, the term
      "description": "<string>",  // required, definition of the term
      "synonyms": ["<string>"]    // optional, alternative names
    }
  ]}

Example (glossary.json):
  {"terms": [
    {"name": "APR", "description": "Annual Percentage Rate", "synonyms": ["annual rate"]},
    {"name": "KYC", "description": "Know Your Customer"}
  ]}

Examples:
  iai glossaries update finance-terms --file glossary.json
  iai glossaries update finance-terms --file glossary.json --labels production,staging`,
		DeleteLong: `Delete a glossary definition and all its versions, or delete specific versions.

Without flags, deletes the glossary entry and all its versions (requires
confirmation). Use --version to delete a specific version, or --label to delete
versions with a specific label. Use -f to skip the confirmation prompt.

Examples:
  iai glossaries delete finance-terms
  iai glossaries delete finance-terms -f
  iai glossaries delete finance-terms --version 3
  iai glossaries delete finance-terms --label staging`,
	})
}
