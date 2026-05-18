## iai skills create

Create a skill

### Synopsis

Create a new Copilot skill for the interactive-chat service.

NOTE: This manages Copilot skills only — it does NOT affect the conversational
agent (interactive-agent). For agent behaviors use 'iai routines', 'iai policies',
'iai glossaries', or 'iai macros'.

The skill body is markdown — either a file path via --file (recommended for
multi-line content) or inline text via --body for one-liners. Optional
--description and --intents populate the config.skill block consumed by the
Copilot runtime to assemble its intent → skill routing table.

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
  iai skills create greet --body "Say hello to the user." \
    --description "Greet the user"
  iai skills create summarize-trace --file ./skill.md --labels production

```
iai skills create <name> [flags]
```

### Options

```
      --body string           Prompt content provided inline (alternative to --file)
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

* [iai skills](iai_skills.md)	 - Manage Copilot (interactive-chat) skills — NOT interactive-agent behaviors

