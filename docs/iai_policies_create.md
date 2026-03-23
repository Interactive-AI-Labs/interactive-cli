## iai policies create

Create a policy

### Synopsis

Create a new policy in an InteractiveAI project.

Content is provided via a YAML file using the --file flag and must follow the
policy schema below. Use --skip-schema to bypass validation.


### Schema

```yaml
policies:                                    # required, array of policy rules
  - id: <string>                             # required, unique identifier
    condition: <string>                      # required, when this rule applies
    action: <string>                         # required, what the agent should do
    criticality: <HIGH|MEDIUM|LOW>           # optional, default MEDIUM
    description: <string>                    # optional
    tools: [<string>, ...]                   # optional, tools to use
    prioritize_over: [<id>, ...]             # optional, policy IDs this overrides
```

### Example

```yaml
# policy.yaml
policies:
  - id: p1
    condition: user requests account deletion
    action: confirm identity before proceeding
    criticality: HIGH
```

### Options

```
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --skip-schema           Skip schema validation (allows draft/WIP content)
      --tags strings          Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional resource definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai policies](iai_policies.md)	 - Manage policies

