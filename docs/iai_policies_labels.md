## iai policies labels

Manage labels on policies versions

### Synopsis

Manage labels on existing policies versions.

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

* [iai policies](iai_policies.md)	 - Single-step behavioral rules for agents
* [iai policies labels set](iai_policies_labels_set.md)	 - Set labels on a policy version

