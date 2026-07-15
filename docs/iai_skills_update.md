## iai skills update

Update a skill (creates a new version)

### Synopsis

Update a Copilot skill by creating a new version with updated content.

Each update creates a brand-new version with exactly the content and config
provided on the command line — the previous version is preserved unchanged
but is not inherited from. In particular, if --description or --intents are
omitted the new version's config.skill block will be empty, even if the
prior version had values for them. Pass them again on every update if you
want the new version to keep them.

Pass --intents once per intent (the flag is repeatable).

```
iai skills update <name> [flags]
```

### Examples

```
  iai skills update summarize-trace --file ./skill.md \
    --description "Summarize a Langfuse trace" \
    --intents "summarize trace" --intents "explain trace"
  iai skills update summarize-trace --file ./skill.md --labels active
```

### Options

```
      --description string      Short description of the skill (stored in config.skill.description)
      --file string             Path to the file containing the updated prompt content
  -h, --help                    help for update
      --intents stringArray     Natural-language intent that triggers this skill — repeat the flag once per intent (stored in config.skill.intents)
      --labels strings          Labels for the new prompt version (comma-separated)
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name that owns the prompts
      --schema-version string   Schema version to validate against (defaults to latest stable)
      --tags strings            Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai skills](iai_skills.md)	 - Manage Interactive Copilot skills (not to be confused with context items that configure the Interactive Agent)

