## iai glossaries update

Update a glossary (creates a new version)

### Synopsis

Update a glossary definition by creating a new version with updated content.

This creates a new version of the glossary using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.


Run `iai glossaries schema` to see the current field definitions.

### Example

```json
{
  "terms": {
    "aht": {
      "name": "AHT",
      "description": "Average Handle Time",
      "synonyms": ["handle time"]
    }
  }
}
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

* [iai glossaries](iai_glossaries.md)	 - Manage glossary definitions

