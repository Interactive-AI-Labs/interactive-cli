## iai stacks

Declarative resource sync from config files

### Synopsis

Manage stacks and their resources (services, agents, databases) from stack configuration files.

### Options

```
  -h, --help   help for stacks
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai stacks diff](iai_stacks_diff.md)	 - Show differences between local config and live stack
* [iai stacks get](iai_stacks_get.md)	 - Export live stack configuration as a stack YAML file
* [iai stacks list](iai_stacks_list.md)	 - List stacks in a project
* [iai stacks sync](iai_stacks_sync.md)	 - Sync services, agents, and databases from a stack config file

