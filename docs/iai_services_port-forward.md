## iai services port-forward

Forward a local port to a service

### Synopsis

Open a local TCP listener and tunnel traffic through the deployment operator
to a service running in the cluster.

The remote port defaults to the service's configured port. Use --port to
override. Use --local-port to choose the local listening port (defaults to
--port when set, or an available OS-assigned port otherwise).

```
iai services port-forward <service_name> [flags]
```

### Examples

```
  iai services port-forward my-svc
  iai services port-forward my-svc --port 8080
  iai services port-forward my-svc --port 8080 --local-port 9090
```

### Options

```
  -h, --help                  help for port-forward
      --local-port int        Local port to listen on (defaults to the remote port)
  -o, --organization string   Organization name
      --port int              Remote port on the service (defaults to the service's configured port)
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

* [iai services](iai_services.md)	 - Deploy and manage HTTP services

