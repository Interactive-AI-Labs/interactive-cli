## iai skills create

Create a skill

### Synopsis

Create a new Copilot skill for the interactive-copilot service.

The skill body is provided as markdown via --file. Optional --description and
--intents populate the config.skill block consumed by the Copilot runtime to
assemble its intent → skill routing table.

Pass --intents once per intent; the flag is repeatable so individual
intents may contain commas (e.g. "summarize, then explain").

Example (skill.md):
  # Summarize Trace

  Given a Langfuse trace ID, fetch the trace and summarize key steps,
  latencies, and any errors.

The server automatically assigns the "latest" label to new versions. To make
a version retrievable via the default 'get' (which resolves "production"),
assign the "production" label with --labels production.

Examples:
  iai skills create summarize-trace --file ./skill.md \
    --description "Summarize a Langfuse trace" \
    --intents "summarize trace" --intents "explain trace"
  iai skills create summarize-trace --file ./skill.md --labels production

```
iai skills create <name> [flags]
```

### Options

```
      --description string    Short description of the skill (stored in config.skill.description)
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --intents stringArray   Natural-language intent that triggers this skill — repeat the flag once per intent (stored in config.skill.intents)
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

* [iai skills](iai_skills.md)	 - Manage Copilot (interactive-copilot) skills — NOT interactive-agent behaviors

