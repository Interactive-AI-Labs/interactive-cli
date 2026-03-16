## iai macros create

Create a macro

### Synopsis

Create a new macro in an InteractiveAI project.

Content is provided via a text or markdown file using the --file flag.

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai macros create disclaimer --file disclaimer.md
  iai macros create disclaimer --file disclaimer.md --labels production
  iai macros create disclaimer --file disclaimer.md --tags legal

```
iai macros create <name> [flags]
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

* [iai macros](iai_macros.md)	 - Manage macros

