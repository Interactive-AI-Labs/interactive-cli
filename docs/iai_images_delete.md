## iai images delete

Delete an image from a project

### Synopsis

Delete the image version identified by its name and tag from a project.
Other tags pointing to the same version are also deleted.

```
iai images delete <image_name> [flags]
```

### Examples

```
  iai images delete my-service --tag 1.2.3
  iai images delete my-service --tag 1.2.3 --force
  iai images delete my-service --tag 1.2.3 --organization my-org --project my-project
```

### Options

```
  -f, --force                 Skip confirmation prompt
  -h, --help                  help for delete
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name the image belongs to
  -t, --tag string            Tag of the image version to delete
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

