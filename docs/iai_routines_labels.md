## iai routines labels

Manage labels on routines versions

### Synopsis

Manage labels on existing routines versions.

This command group reassigns labels to existing versions without creating a
new version. Labels are unique per prompt: assigning a label to one version
removes it from any other version that currently has it.

### Options

```
  -h, --help   help for labels
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai routines](iai_routines.md)	 - Multi-step behavioral processes for agents
* [iai routines labels set](iai_routines_labels_set.md)	 - Set labels on a routine version

