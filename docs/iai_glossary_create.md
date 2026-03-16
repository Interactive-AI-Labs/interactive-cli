## iai glossary create

Create a glossary

### Synopsis

Create a new glossary definition in an InteractiveAI project.

Content is provided via a JSON file using the --file flag and must follow the
glossary schema (see 'iai glossary --help'). Use --skip-schema to bypass validation.

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai glossary create finance-terms --file glossary.json
  iai glossary create finance-terms --file glossary.json --labels production
  iai glossary create finance-terms --file glossary.json --tags domain --skip-schema

```
iai glossary create <name> [flags]
```

### Options

```
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --skip-schema           Skip schema validation (allows draft/WIP content)
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

* [iai glossary](iai_glossary.md)	 - Manage glossary definitions

