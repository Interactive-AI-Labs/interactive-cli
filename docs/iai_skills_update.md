## iai skills update

Update a skill (creates a new version)

### Synopsis

Update a skill by creating a new version with updated content.

This creates a new version of the skill using the content from the provided
file or --body string. --description and --intents are written into the new
version's config.skill block; omit them to leave the previous version's
config alone (they are stored per version).

Examples:
  iai skills update summarize-trace --file ./skill.md
  iai skills update summarize-trace --body "# Greet\n\nSay hello." \
    --description "Greet the user" --intents "say hi,greet"
  iai skills update summarize-trace --file ./skill.md --labels production,staging

```
iai skills update <name> [flags]
```

### Options

```
      --body string           Prompt content provided inline (alternative to --file)
      --description string    Short description of the skill (stored in config.skill.description)
      --file string           Path to the file containing the updated prompt content
  -h, --help                  help for update
      --intents strings       Natural-language intents that trigger this skill (comma-separated; stored in config.skill.intents)
      --labels strings        Labels for the new prompt version (comma-separated)
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

* [iai skills](iai_skills.md)	 - Copilot skills loaded by interactive-chat at runtime

