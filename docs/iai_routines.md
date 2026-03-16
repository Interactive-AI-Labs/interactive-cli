## iai routines

Manage routines

### Synopsis

Manage routines in InteractiveAI projects.

Routines are step-by-step conversation flows with branching logic and terminal
states (YAML format).


### Schema

```yaml
steps:                                       # required, array of steps
  - step: <string>                           # required, step identifier
    name: <string>                           # required, step display name
    type: <node|branch|finish|branchnode>    # required
    description: <string>                    # optional
    tool: <string>                           # optional, tool to invoke
    condition: <string>                      # optional, branching condition
    input: <string>                          # optional
    output: <string>                         # optional
```

### Example

```yaml
# routine.yaml
steps:
  - step: "1"
    name: Greet
    type: node
    description: "Welcome the user"
  - step: "2"
    name: Done
    type: finish
```

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
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai routines create](iai_routines_create.md)	 - Create a routine
* [iai routines delete](iai_routines_delete.md)	 - Delete a routine
* [iai routines get](iai_routines_get.md)	 - Get details of a routine
* [iai routines list](iai_routines_list.md)	 - List routines in a project
* [iai routines update](iai_routines_update.md)	 - Update a routine (creates a new version)

