## Structured output

Many commands support `--json` and `--yaml` for automation. Human-readable table/detail output is the default.

For API-backed commands such as `traces`, `observations`, `datasets`, `queues`, `scores`, and `sessions`, structured output preserves the API envelope:

```json
{
  "success": true,
  "data": {
    "traces": []
  }
}
```

Use the resource under `data` as the stable payload, for example:

```bash
iai traces list --json | jq '.data.traces[]'
iai traces get <trace-id> --yaml
```

Prompt resources (`prompts`, `routines`, `policies`, `variables`, `glossaries`, `macros`, `skills`) expose structured output on `list` and `describe/get`. List output is wrapped as:

```json
{
  "prompts": [],
  "totalCount": 0
}
```

