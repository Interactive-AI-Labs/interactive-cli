## iai agents select

Select an agent for subsequent commands

### Synopsis

Select an agent by name and store it in the local CLI configuration so other
commands can use it without specifying the agent each time. Commands that create,
change or remove an agent still take the name. Selecting another organization or
project clears the selected agent.

```
iai agents select [agent_name] [flags]
```

### Examples

```
  iai agents select my-agent
  iai agents select --clear
```

### Options

```
      --clear                 Clear the selected agent
  -h, --help                  help for select
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

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools

