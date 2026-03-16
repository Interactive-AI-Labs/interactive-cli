## iai policies

Manage policies

### Synopsis

Manage policies in InteractiveAI projects.

Policies are core behavior rules — condition-action pairs that govern agent
responses (YAML format).


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
  -h, --help   help for policies
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai policies create](iai_policies_create.md)	 - Create a policy
* [iai policies delete](iai_policies_delete.md)	 - Delete a policy
* [iai policies get](iai_policies_get.md)	 - Get details of a policy
* [iai policies list](iai_policies_list.md)	 - List policies in a project
* [iai policies update](iai_policies_update.md)	 - Update a policy (creates a new version)

