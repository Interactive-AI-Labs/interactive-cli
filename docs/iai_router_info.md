## iai router info

Display router endpoint information

### Synopsis

Display the inference router base URL, endpoints, and documentation URL.

```
iai router info [flags]
```

### Examples

```
  iai router info
  iai router info --json
  iai router info --yaml
  iai router info --hostname https://dev.interactive.ai
```

### Options

```
  -h, --help   help for info
      --json   Output router information as JSON
      --yaml   Output router information as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai router](iai_router.md)	 - Inspect the inference router and its models

