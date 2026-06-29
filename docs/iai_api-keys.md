## iai api-keys

Project API keys

### Synopsis

Manage project API keys. Requires iai login or JWT authentication. API key authentication is not supported.

Project API keys authenticate platform/API access for reading and writing project context, such as prompts, routines, policies, variables, glossaries, macros, traces, scores, datasets, and for creating infrastructure resources such as agents, services, and databases.

### Options

```
  -h, --help   help for api-keys
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
* [iai api-keys create](iai_api-keys_create.md)	 - Create a project API key
* [iai api-keys delete](iai_api-keys_delete.md)	 - Delete a project API key
* [iai api-keys list](iai_api-keys_list.md)	 - List project API keys

