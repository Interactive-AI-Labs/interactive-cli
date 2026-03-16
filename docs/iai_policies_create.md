## iai policies create

Create a policy

### Synopsis

Create a new policy in an InteractiveAI project.

Content is provided via a YAML file using the --file flag and must follow the
policy schema (see 'iai policies --help'). Use --skip-schema to bypass validation.

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai policies create safety-rules --file policy.yaml
  iai policies create safety-rules --file policy.yaml --labels production
  iai policies create safety-rules --file policy.yaml --tags compliance --skip-schema

```
iai policies create <name> [flags]
```

### Options

```
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --skip-schema           Skip schema validation (allows draft/WIP content)
      --tags strings          Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai policies](iai_policies.md)	 - Manage policies

