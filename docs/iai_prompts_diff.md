## iai prompts diff

Compare two versions of a prompt

### Synopsis

Show the differences between two versions of a prompt.

```
iai prompts diff <name> <version_a> <version_b> [flags]
```

### Examples

```
  iai prompts diff greeting 1 3
```

### Options

```
  -h, --help                  help for diff
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

* [iai prompts](iai_prompts.md)	 - Versioned prompts for agents, evaluators, and guardrails

