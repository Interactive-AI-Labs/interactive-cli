Run `iai glossaries schema` to see the current field definitions.

### Example

```json
{
  "terms": {
    "aht": {
      "name": "AHT",
      "description": "Average Handle Time",
      "synonyms": ["handle time"]
    },
    "kyc": {
      "name": "KYC",
      "description": "Know Your Customer",
      "synonyms": ["identity check"]
    }
  }
}
```

Each key under `terms` must be unique. Add as many terms as you need — the
glossary accepts any number of entries.
