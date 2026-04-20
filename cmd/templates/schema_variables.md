Run `iai variables schema` to see the current field definitions.

### Example

```json
{
  "variables": {
    "user_name": {
      "description": "The user's display name",
      "default_value": "Guest"
    },
    "is_authenticated": {
      "description": "Whether the user has completed sign-in",
      "default_value": false
    }
  }
}
```

Each key under `variables` must be unique. Add as many entries as you need —
the set accepts any number of variables.
