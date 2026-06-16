## iai agents port-forward

Forward a local port to an agent

### Synopsis

Open a local TCP listener and tunnel traffic through the deployment operator
to an agent running in the cluster.

The remote port defaults to the agent's configured port. Use --port to
override. Use --local-port to choose the local listening port (defaults to
--port when set, or an available OS-assigned port otherwise).

```
iai agents port-forward <agent_name> [flags]
```

### Examples

```
  iai agents port-forward my-agent
  iai agents port-forward my-agent --port 8080
  iai agents port-forward my-agent --port 8080 --local-port 9090
```

### Options

```
  -h, --help                  help for port-forward
      --local-port int        Local port to listen on (defaults to the remote port)
  -o, --organization string   Organization name
      --port int              Remote port on the agent (defaults to the agent's configured port)
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

