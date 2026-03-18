### Schema

```json
{
  "variables": [
    {
      "name": "<string>",
      "type": "<boolean|string|number|array|object>",
      "persistence": "<session|customer|global>",
      "default_value": "<any>"
    }
  ]
}
```

> `name` and `type` are required. `persistence` defaults to `"session"`. `default_value` is optional.

### Example

```json
{
  "variables": [
    {"name": "user_name", "type": "string"},
    {"name": "is_authenticated", "type": "boolean", "default_value": false},
    {"name": "preferences", "type": "object", "persistence": "customer"}
  ]
}
```
