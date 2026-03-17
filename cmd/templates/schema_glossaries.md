### Schema

```json
{
  "terms": [
    {
      "name": "<string>",
      "description": "<string>",
      "synonyms": ["<string>"]
    }
  ]
}
```

> `name` and `description` are required. `synonyms` is optional.

### Example

```json
{
  "terms": [
    {"name": "APR", "description": "Annual Percentage Rate", "synonyms": ["annual rate"]},
    {"name": "KYC", "description": "Know Your Customer"}
  ]
}
```
