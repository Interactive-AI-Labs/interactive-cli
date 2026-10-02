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

