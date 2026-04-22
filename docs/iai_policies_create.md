## iai policies create

Create a policy

### Synopsis

Create a new policy in an InteractiveAI project.

Each prompt holds exactly one policy (flat YAML, fields at the root).
Content is provided via a YAML file using the --file flag.

Run `iai policies schema` to see the current field definitions.

### Example

```yaml
# policy.yaml
id: escalate
name: Escalation Policy
condition: User requests human agent
action: Transfer to human
criticality: HIGH
# always_match: true  # evaluate on every turn regardless of context
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

* [iai policies](iai_policies.md)	 - Manage policies

