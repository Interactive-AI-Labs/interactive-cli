## iai images build

Build a container image with Docker

### Synopsis

Build a container image using the local Docker CLI.

This is a thin wrapper around 'docker build' that requires an explicit tag,
Dockerfile, and build context.

```
iai images build [image_name] [flags]
```

### Examples

```
  iai images build my-service --tag 1.2.3
  iai images build my-service --tag 1.2.3 --file docker/Dockerfile --context .
  iai images build my-service --tag 1.2.3 --platform linux/amd64
```

### Options

```
  -c, --context string    Build context directory (default: current directory) (default ".")
  -f, --file string       Path to the Dockerfile (default: ./Dockerfile) (default "Dockerfile")
  -h, --help              help for build
      --platform string   Target platform for the build (currently only linux/amd64 is supported) (default "linux/amd64")
  -t, --tag string        Tag suffix to append to the fixed registry (e.g. 1.2.3)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai images](iai_images.md)	 - Manage container images

