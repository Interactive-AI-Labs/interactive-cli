## iai stacks sync

Sync services, agents, databases, mcps, and jobs from a stack config file

### Synopsis

Sync services, agents, databases, mcps, and jobs in a project from a stack configuration file.

Services, agents, databases, mcps, and jobs are created and updated to match the config
file. Resources the config file no longer mentions are NOT deleted by
default: a config that omits a resource looks identical to a stale one, so
the sync refuses each deletion, reports it on stderr, and continues with the
creates and updates. Pass --allow-delete with the resource types you intend
to decommission (services, agents, databases, mcps, jobs, or all) to delete them;
within each resource type, deletes run after that type's creates and updates.

Resource types sync in order: services, databases, mcps, then agents once
the self-hosted mcps are ready, and finally jobs. Remote mcps are registered
on the platform.

Updates replace the whole live spec of each resource. For every service, agent,
mcp, or job that changes, the live revision being replaced is printed to stderr
so a sync from a stale config file is visible. Jobs can only be updated or
deleted when all their runs have finished.

Script jobs reference their files with scriptFile and pyprojectFile, resolved
relative to the config file.

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
  iai stacks sync --file stack.yaml --wait-timeout 10m
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

> **Note:** a remote mcp takes exactly one of `catalogId` or `endpointUrl`, an `auth.type` of none, bearer, api_key or custom, and none of the self-hosted keys; a catalog entry supplies its own endpoint and auth headers. Create oauth and client_credentials mcps with `iai mcps create`; a stack never manages them.

```yaml
mcps:
  tools:
    type: self-hosted
    port: 8080
    image:
      type: internal
      name: my-mcp
      tag: v1
    resources:
      memory: "128M"
      cpu: "250m"
    auth:
      type: none
  docs:
    type: remote
    catalogId: awsknowledge
    auth:
      type: none
```


### Options

```
      --allow-delete strings    Resource types the sync may delete when the config omits them (services, agents, databases, mcps, jobs, or all); deletions are refused otherwise
      --dry-run                 Print the full plan (creates, updates, deletes, refused deletions) without applying anything
  -f, --file string             Path to stack configuration file
  -h, --help                    help for sync
      --no-wait                 Sync agents without waiting for the self-hosted mcps to be ready
  -o, --organization string     Organization name that owns the project
  -p, --project string          Project name to sync resources in
      --wait-timeout duration   How long to wait for every self-hosted mcp in the config to be ready (tools verified) before syncing agents; if one is not ready in time the sync fails and agents are left unchanged (default 5m0s)
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

