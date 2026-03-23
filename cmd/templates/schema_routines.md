Run `iai routines schema` to see the current field definitions.

### Example

```yaml
# routine.yaml
title: My Routine
conditions: When user needs help
description: Handles user support requests
steps:
  - id: greet
    description: Welcome the user
    chat_state: Say hello
  - id: lookup
    source: greet
    tools: crm:get_user
    tool_instruction: Fetch user data
```
