## iai glossaries create

Create a glossary

### Synopsis

Create a new glossary definition in an InteractiveAI project.

Content is provided via a JSON file using the --file flag.

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
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --tags strings          Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai glossaries](iai_glossaries.md)	 - Manage glossary definitions

