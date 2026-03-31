## iai prompts create

Create a prompt

### Synopsis

Create a new text or chat prompt in an InteractiveAI project.

Content is provided via --file (path to a file) or --content (inline string).
Exactly one of --file or --content must be specified.

The --type flag selects the prompt type: "text" (default) or "chat".

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

Examples:
  iai prompts create greeting --content "Hello, how can I help you?"
  iai prompts create greeting --file greeting.txt
  iai prompts create greeting --file greeting.txt --type chat
  iai prompts create greeting --content "Hi!" --labels production
  iai prompts create greeting --file greeting.txt --tags support

```
iai prompts create <name> [flags]
```

### Options

```
      --content string        Inline prompt content string
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --tags strings          Tags for the prompt (comma-separated)
      --type string           Prompt type: "text" (default) or "chat" (default "text")
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai prompts](iai_prompts.md)	 - Manage prompts

