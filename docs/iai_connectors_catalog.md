## iai connectors catalog

Browse the connector catalog

### Synopsis

List the curated catalog of MCP servers you can connect to with
'iai connectors create --catalog-id', showing each entry's id, category, and
supported auth methods.

```
iai connectors catalog [flags]
```

### Examples

```
  iai connectors catalog
  iai connectors catalog --json
```

### Options

```
  -h, --help   help for catalog
      --json   Output raw API response as JSON
      --yaml   Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the connectors
```

### SEE ALSO

* [iai connectors](iai_connectors.md)	 - Manage MCP connectors in a project

