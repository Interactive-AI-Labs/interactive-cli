## iai router keys

Router API keys

### Synopsis

Manage InteractiveAI Router API keys. Requires iai login or JWT authentication. API key authentication is not supported.

Router keys authenticate inference requests to the InteractiveAI Router, for example chat completions and model calls. They are used as bearer tokens for runtime inference, not for managing project context or infrastructure.

### Options

```
  -h, --help   help for keys
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai router](iai_router.md)	 - Inspect the inference router, keys, and models
* [iai router keys create](iai_router_keys_create.md)	 - Create a router API key
* [iai router keys delete](iai_router_keys_delete.md)	 - Delete a router API key
* [iai router keys list](iai_router_keys_list.md)	 - List router API keys
* [iai router keys update](iai_router_keys_update.md)	 - Update a router API key

