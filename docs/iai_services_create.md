## iai services create

Create a service in a project

### Synopsis

Create a service in a specific project using the deployment service.

```
iai services create <service_name> [flags]
```

### Examples

```
  iai services create my-svc --image-type external --image-repository docker.io --image-name nginx --image-tag latest --port 80 --memory 512M --cpu 0.5
  iai services create my-svc --image-name my-app --image-tag v1 --port 8080 --memory 1G --cpu 1 --replicas 3 --endpoint
  iai services create my-svc --image-name my-app --image-tag v1 --memory 512M --cpu 0.5 --env LOG_LEVEL=debug --secret DB_PASSWORD --healthcheck-path /health
  iai services create my-svc --image-name my-app --image-tag v1 --memory 512M --cpu 0.5 --schedule-uptime "Mon-Fri 08:00-18:00" --schedule-timezone Europe/Berlin
```

### Options

```
      --autoscaling-cpu-percentage int      CPU percentage threshold for autoscaling
      --autoscaling-max-replicas int        Maximum number of replicas for autoscaling
      --autoscaling-memory-percentage int   Memory percentage threshold for autoscaling
      --autoscaling-min-replicas int        Minimum number of replicas for autoscaling
      --cpu string                          CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)
      --endpoint                            Expose the service at <service-name>-<project-hash>.interactive.ai
      --env stringArray                     Environment variable (NAME=VALUE); can be repeated
      --healthcheck-initial-delay int       Initial delay in seconds before starting healthchecks
      --healthcheck-path string             HTTP path for healthcheck endpoint (e.g. /health)
  -h, --help                                help for create
      --image-name string                   Container image name
      --image-repository string             Container image repository (external images only)
      --image-tag string                    Container image tag
      --image-type string                   Image type: 'internal' (project's private registry), 'external' (any public registry), or 'platform' (Interactive AI registries)
      --memory string                       Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)
  -o, --organization string                 Organization name that owns the project
      --port int                            Service port to expose
  -p, --project string                      Project name to create the service in
      --replicas int                        Number of replicas for the service (mutually exclusive with autoscaling)
      --schedule-downtime string            When the service should be scaled down (mutually exclusive with --schedule-uptime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Weekdays: Mon, Tue, Wed, Thu, Fri, Sat, Sun (case-insensitive). Times in 24h format; start: 00:00-23:59, end: 00:00-24:00 (24:00 = end of day). Example: 'Sat-Sun 00:00-24:00'
      --schedule-timezone string            IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); required with --schedule-uptime or --schedule-downtime
      --schedule-uptime string              When the service should be running (mutually exclusive with --schedule-downtime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Weekdays: Mon, Tue, Wed, Thu, Fri, Sat, Sun (case-insensitive). Times in 24h format; start: 00:00-23:59, end: 00:00-24:00 (24:00 = end of day). Example: 'Mon-Fri 07:30-20:30' or 'Mon-Fri 08:00-18:00, Sat 10:00-14:00'
      --secret stringArray                  Secrets to be loaded as env vars; can be repeated
      --stack-id string                     Stack ID to assign the service to
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

