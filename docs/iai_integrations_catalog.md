## iai integrations catalog

Browse the MCP integrations catalog

### Synopsis

List the curated catalog of MCP servers you can connect to with
'iai integrations create-from-catalog', showing each entry's id, category, and
supported auth methods.

Examples:
  iai integrations catalog

```
iai integrations catalog [flags]
```

### Options

```
  -h, --help                  help for catalog
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name to browse the catalog for
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai integrations](iai_integrations.md)	 - MCP integration connections for a project

