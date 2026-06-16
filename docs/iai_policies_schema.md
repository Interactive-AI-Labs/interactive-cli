## iai policies schema

Display the JSON Schema for policies

### Synopsis

Fetch and display the JSON Schema for policies from the backend API.

Use --schema-version to request a specific schema version (defaults to latest stable).
Use --json or --yaml to output the schema response in a structured format.

This is a public endpoint and does not require authentication.

```
iai policies schema [flags]
```

### Examples

```
  iai policies schema
  iai policies schema --schema-version 0.0.1
  iai policies schema --json
```

### Options

```
  -h, --help                    help for schema
      --json                    Output schema response as JSON
      --schema-version string   Schema version to fetch (defaults to latest stable)
      --yaml                    Output schema response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai policies](iai_policies.md)	 - Single-step behavioral rules for agents

