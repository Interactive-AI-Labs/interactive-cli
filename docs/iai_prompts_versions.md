## iai prompts versions

List versions of a prompt

### Synopsis

List all versions of a prompt, sorted newest-first.

Each row shows the version number, when it was updated, who updated it, and the
commit message recorded with that version (set with -m on create and update).
A "—" means the value is unavailable: under API-key authentication only version
numbers can be read.

```
iai prompts versions <name> [flags]
```

### Examples

```
  iai prompts versions greeting
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

* [iai prompts](iai_prompts.md)	 - Versioned prompts for agents, evaluators, and guardrails

