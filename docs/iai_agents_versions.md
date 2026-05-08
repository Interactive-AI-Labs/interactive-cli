## iai agents versions

List versions of an agent

### Synopsis

Show past versions of an agent, sorted newest-first.
Up to 50 versions are retained per agent.

Examples:
  iai agents versions my-agent

```
iai agents versions <agent_name> [flags]
```

### Options

```
  -h, --help                  help for versions
  -o, --organization string   Organization name
  -p, --project string        Project name
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Manage agents

