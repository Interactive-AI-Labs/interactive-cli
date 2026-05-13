## iai policies versions

List versions of a policy

### Synopsis

List all versions of a policy, sorted newest-first.

Examples:
  iai policies versions my-policy

```
iai policies versions <name> [flags]
```

### Options

```
  -h, --help                  help for versions
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
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

