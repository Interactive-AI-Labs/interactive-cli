## iai policies create

Create a policy

### Synopsis

Create a new policy in an InteractiveAI project.

Content is provided via a YAML file using the --file flag.

Run `iai policies schema` to see the current field definitions.

### Example

```yaml
# policy.yaml
policies:
  - id: escalate
    name: Escalation Policy
    condition: User requests human agent
    action: Transfer to human
    criticality: HIGH
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

* [iai policies](iai_policies.md)	 - Manage policies

