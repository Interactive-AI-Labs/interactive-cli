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

