### Schema

```yaml
steps:                                       # required, array of steps
  - step: <string>                           # required, step identifier
    name: <string>                           # required, step display name
    type: <node|branch|finish|branchnode>    # required
    description: <string>                    # optional
    tool: <string>                           # optional, tool to invoke
    condition: <string>                      # optional, branching condition
    input: <string>                          # optional
    output: <string>                         # optional
```

### Example

```yaml
# routine.yaml
steps:
  - step: "1"
    name: Greet
    type: node
    description: "Welcome the user"
  - step: "2"
    name: Done
    type: finish
```
