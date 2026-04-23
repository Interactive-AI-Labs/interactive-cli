Run `iai policies schema` to see the current field definitions.

### Example

```yaml
# policy.yaml
id: escalate
name: Escalation Policy
condition: User requests human agent
action: Transfer to human
criticality: HIGH
# Uncomment to evaluate on every turn regardless of context match:
# always_match: true
```
