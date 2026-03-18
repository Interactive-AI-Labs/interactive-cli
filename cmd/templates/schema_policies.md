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
