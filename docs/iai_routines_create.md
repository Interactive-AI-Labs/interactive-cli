## iai routines create

Create a routine

### Synopsis

Create a new routine in an InteractiveAI project.

Content is provided via a YAML file using the --file flag.

Run `iai routines schema` to see the current field definitions.

### Example

```yaml
# routine.yaml
title: My Routine
conditions: When user needs help
description: Handles user support requests
steps:
  - id: greet
    description: Welcome the user
    chat_state: Say hello
  - id: lookup
    source: greet
    tools: crm:get_user
    tool_instruction: Fetch user data
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

* [iai routines](iai_routines.md)	 - Manage routines

