## iai stacks sync

Sync services, agents, databases, and mcps from a stack config file

### Synopsis

Sync services, agents, databases, and mcps in a project from a stack configuration file.

Services, agents, databases, and mcps are created and updated to match the config
file. Resources the config file no longer mentions are NOT deleted by
default: a config that omits a resource looks identical to a stale one, so
the sync refuses each deletion, reports it on stderr, and continues with the
creates and updates. Pass --allow-delete with the resource types you intend
to decommission (services, agents, databases, mcps, or all) to delete them; within
each resource type, deletes run after that type's creates and updates.

Updates replace the whole live spec of each resource. For every service, agent,
or mcp updated, the live revision being replaced is printed to stderr so a
sync from a stale config file is visible before it lands.

Use --dry-run to print the full plan — creates, updates, deletes, and
refused deletions — without applying anything.

The organization and project are read from the config file, flags, or resolved via 'iai organizations select' / 'iai projects select'.

```
iai stacks sync [flags]
```

### Examples

```
  iai stacks sync --file stack.yaml
  iai stacks sync --file stack.yaml --project my-project --organization my-org
  iai stacks sync --file stack.yaml --dry-run
  iai stacks sync --file stack.yaml --allow-delete services,agents
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
      --allow-delete strings   Resource types the sync may delete when the config omits them (services, agents, databases, mcps, or all); deletions are refused otherwise
      --dry-run                Print the full plan (creates, updates, deletes, refused deletions) without applying anything
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

* [iai stacks](iai_stacks.md)	 - Declarative resource sync from config files

