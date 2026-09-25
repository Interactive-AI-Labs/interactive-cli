## iai secrets

Encrypted key-value pairs for services and agents

### Synopsis

Manage secrets in InteractiveAI projects.

Names starting with iai-mcp- and keys starting with IAI_MCP_ are reserved for
MCP credentials. Those are set and rotated with 'iai mcps create|update
--credential', aren't listed here, and can't be read, changed or deleted here.

### Options

```
  -h, --help   help for secrets
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai secrets create](iai_secrets_create.md)	 - Create a secret in a project
* [iai secrets delete](iai_secrets_delete.md)	 - Delete a secret in a project
* [iai secrets get](iai_secrets_get.md)	 - Get a secret in a project
* [iai secrets list](iai_secrets_list.md)	 - List secrets in a project
* [iai secrets update](iai_secrets_update.md)	 - Update keys in a secret

