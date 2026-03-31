## iai variables create

Create a variable

### Synopsis

Create a new variable definition in an InteractiveAI project.

Content is provided via a JSON file using the --file flag.

Run `iai variables schema` to see the current field definitions.

### Example

```json
{
  "variables": {
    "user_name": {
      "description": "The user's display name",
      "default_value": "Guest"
    }
  }
}
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
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai variables](iai_variables.md)	 - Manage variables

