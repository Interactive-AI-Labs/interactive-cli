## iai skills delete

Delete a skill

### Synopsis

Delete a Copilot skill (interactive-chat only — not interactive-agent) and all its versions, or delete specific versions.

Without flags, deletes the skill and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions
with a specific label. Use -f to skip the confirmation prompt.

Examples:
  iai skills delete summarize-trace
  iai skills delete summarize-trace -f
  iai skills delete summarize-trace --version 3
  iai skills delete summarize-trace --label staging

```
iai skills delete <name> [flags]
```

### Options

```
  -f, --force                 Skip confirmation prompt
  -h, --help                  help for delete
      --label string          Delete versions with this label only
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Delete a specific version only
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

