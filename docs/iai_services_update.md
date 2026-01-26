## iai services update

Update a service in a project

### Synopsis

Update a service in a specific project using the deployment service.

All configuration is provided via flags. The project is selected with --project or via 'iai projects select'.

```
iai services update [service_name] [flags]
```

### Options

```
      --autoscaling-cpu-percentage int      CPU percentage threshold for autoscaling
      --autoscaling-enabled                 Enable autoscaling (mutually exclusive with replicas)
      --autoscaling-max-replicas int        Maximum number of replicas when autoscaling is enabled
      --autoscaling-memory-percentage int   Memory percentage threshold for autoscaling
      --autoscaling-min-replicas int        Minimum number of replicas when autoscaling is enabled
      --cpu string                          CPU cores or millicores (e.g. 1, 2, 500m, 1000m) - required
      --endpoint                            Expose the service at <service-name>-<project-hash>.interactive.ai
      --env stringArray                     Environment variable (NAME=VALUE); can be repeated
  -h, --help                                help for update
      --image-name string                   Container image name
      --image-repository string             Container image repository (external images only)
      --image-tag string                    Container image tag
      --image-type string                   Image type: 'external' (Docker Hub, ghcr.io) or 'internal' (InteractiveAI private registry)
      --memory string                       Memory with M or G unit (e.g. 128M, 512M, 1G) - required
  -o, --organization string                 Organization name that owns the project
      --port int                            Service port to expose
  -p, --project string                      Project name to update the service in
      --replicas int                        Number of replicas for the service (mutually exclusive with autoscaling)
      --secret stringArray                  Secrets to be loaded as env vars; can be repeated
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai services](iai_services.md)	 - Manage services

