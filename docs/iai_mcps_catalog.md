## iai mcps catalog

Browse the MCP catalog

### Synopsis

List the MCPs available to create: self-hosted images from the platform's
registry and remote providers.

Without arguments, lists both. With an id argument, lists the available
versions of a self-hosted image, or shows the remote entry.

```
iai mcps catalog [id] [flags]
```

### Examples

```
  iai mcps catalog
  iai mcps catalog gdrive
  iai mcps catalog --json
```

### Options

```
  -h, --help   help for catalog
      --json   Output the catalog as JSON
      --yaml   Output the catalog as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
  -o, --organization string          Organization name that owns the project
  -p, --project string               Project name that owns the mcps
```

### SEE ALSO

* [iai mcps](iai_mcps.md)	 - Deploy and manage MCP servers

