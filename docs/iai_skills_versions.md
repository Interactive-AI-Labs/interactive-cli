## iai skills versions

List versions of a skill

### Synopsis

List all versions of a skill, sorted newest-first.

```
iai skills versions <name> [flags]
```

### Examples

```
  iai skills versions my-skill
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

* [iai skills](iai_skills.md)	 - Manage Interactive Copilot skills (not to be confused with context items that configure the Interactive Agent)

