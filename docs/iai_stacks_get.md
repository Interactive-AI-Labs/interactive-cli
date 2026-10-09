## iai stacks get

Export live stack configuration

### Synopsis

Fetch the live services, agents, databases, mcps, and jobs for a stack and
write them as a stack configuration file.

Use this to rebase your local stack config on the live state before making
changes. MCP credentials are never exported. Existing self-hosted MCPs keep
omitted credentials when their authentication settings are unchanged.
Supply required credentials when creating MCPs or changing authentication.
Remote custom-auth updates also require a credential.

With --file, each script job's files are written to jobs/<name>/main.py and
jobs/<name>/pyproject.toml next to the config file, overwriting existing
files. Other outputs omit script job files; add scriptFile and pyprojectFile
before syncing.

The organization and project are read from flags or resolved via 'iai
organizations select' / 'iai projects select'.

```
iai stacks get [flags]
```

### Examples

```
  iai stacks get --stack-id my-stack
  iai stacks get --stack-id my-stack -f live-stack.yaml
  iai stacks get --stack-id my-stack -o my-org -p my-project
```

### Options

```
  -f, --file string           Write output to file instead of stdout; cannot combine with --json or --yaml
  -h, --help                  help for get
      --json                  Output as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --stack-id string       Stack ID to export
      --yaml                  Output as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai stacks](iai_stacks.md)	 - Declarative resource sync from config files

