## iai routines

Manage routines

### Synopsis

Manage routines in InteractiveAI projects.

Routines are step-by-step conversation flows with branching logic and terminal
states (YAML format).

### Options

```
  -h, --help   help for routines
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai routines create](iai_routines_create.md)	 - Create a routine
* [iai routines delete](iai_routines_delete.md)	 - Delete a routine
* [iai routines get](iai_routines_get.md)	 - Get details of a routine
* [iai routines list](iai_routines_list.md)	 - List routines in a project
* [iai routines schema](iai_routines_schema.md)	 - Display the JSON Schema for routines
* [iai routines update](iai_routines_update.md)	 - Update a routine (creates a new version)

