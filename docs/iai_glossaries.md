## iai glossaries

Manage glossary definitions

### Synopsis

Manage glossary definitions in InteractiveAI projects.

Glossary entries are domain-specific terms with descriptions and synonyms (JSON
format).

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

### Options

```
  -h, --help   help for glossaries
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai glossaries create](iai_glossaries_create.md)	 - Create a glossary
* [iai glossaries delete](iai_glossaries_delete.md)	 - Delete a glossary
* [iai glossaries get](iai_glossaries_get.md)	 - Get details of a glossary
* [iai glossaries list](iai_glossaries_list.md)	 - List glossaries in a project
* [iai glossaries update](iai_glossaries_update.md)	 - Update a glossary (creates a new version)

