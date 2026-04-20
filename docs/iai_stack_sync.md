## iai stack sync

Sync services, agents, and vector stores from a stack config file

### Synopsis

Sync services, agents, and vector stores in a project from a stack configuration file.

Services are created, updated, or deleted to match the config file.
Agents are created, updated, or deleted to match the config file.
Vector stores are created or deleted (--allow-delete=vector-stores). Updates are not yet supported.

The organization and project are read from the config file, flags, or resolved via 'iai organizations select' / 'iai projects select'.

```
iai stack sync [flags]
```

### Example config file

```yaml
organization: my-org
project: my-project
stack-id: my-stack-v1

services:
  my-service:
    servicePort: 8080
    image:
      type: external
      repository: kennethreitz
      name: httpbin
      tag: latest
    resources:
      memory: "512M"
      cpu: "1"
    env:
      - name: DATABASE_URL
        value: "postgres://db:5432/mydb"
      - name: LOG_LEVEL
        value: "info"
    secretRefs:
      - secretName: my-secret
    endpoint: true
    replicas: 2
    healthcheck:
      path: /health
      initialDelaySeconds: 10
    schedule:
      uptime: "Mon-Fri 07:30-20:30"
      timezone: "Europe/Berlin"

vector-stores:
  knowledge-base:
    resources:
      cpu: 2
      memory: 4
    storage:
      size: 50
      autoResize: true
      autoResizeLimit: 200
    ha: false
    backups: true
```

> **Note:** `replicas` and `autoscaling` are mutually exclusive for services. To use autoscaling instead:

```yaml
    autoscaling:
      minReplicas: 2
      maxReplicas: 10
      cpuPercentage: 80
      memoryPercentage: 85
```


### Options

```
      --allow-delete strings   Resource types to allow deletion for (e.g. vector-stores)
  -f, --file string            Path to stack configuration file
  -h, --help                   help for sync
  -o, --organization string    Organization name that owns the project
  -p, --project string         Project name to sync resources in
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai stack](iai_stack.md)	 - Manage stacks

