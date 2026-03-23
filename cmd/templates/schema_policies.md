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
