## iai variables schema

Display the JSON Schema for variables

### Synopsis

Fetch and display the JSON Schema for variables from the backend API.

Use --schema-version to request a specific schema version (defaults to latest stable).
Use --json or --yaml to output the schema response in a structured format.

This is a public endpoint and does not require authentication.

```
iai variables schema [flags]
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

* [iai variables](iai_variables.md)	 - Contextual attributes referenced in policies and routines

