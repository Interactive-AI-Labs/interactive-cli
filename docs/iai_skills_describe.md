## iai skills describe

Describe a skill in detail

### Synopsis

Show a Copilot skill in detail, including its config and full content.

NOTE: Copilot skills (interactive-chat) only — not interactive-agent behaviors.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai skills describe summarize-trace
  iai skills describe summarize-trace --version 3
  iai skills describe summarize-trace --label staging

```
iai skills describe <name> [flags]
```

### Options

```
  -h, --help                  help for describe
      --label string          Retrieve the version with this label (default: server resolves 'production')
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Retrieve a specific version number
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai skills](iai_skills.md)	 - Manage Copilot (interactive-chat) skills — NOT interactive-agent behaviors

