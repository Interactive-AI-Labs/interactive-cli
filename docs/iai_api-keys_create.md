## iai api-keys create

Create a project API key

### Synopsis

Create a project API key.

Project API keys authenticate platform/API access for reading and writing project context, such as prompts, routines, policies, variables, glossaries, macros, traces, scores, datasets, and for creating infrastructure resources such as agents, services, and databases.

```
iai api-keys create [flags]
```

### Options

```
  -h, --help                  help for create
      --json                  Output response as JSON
      --note string           API key note
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai api-keys](iai_api-keys.md)	 - Project API keys

