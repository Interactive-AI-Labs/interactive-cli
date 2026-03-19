## iai routines create

Create a routine

### Synopsis

Create a new routine in an InteractiveAI project.

Content is provided via a YAML file using the --file flag and must follow the
routine schema below.


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
      --file string           Path to the file containing the prompt content
  -h, --help                  help for create
      --labels strings        Labels for the prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --tags strings          Tags for the prompt (comma-separated)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai routines](iai_routines.md)	 - Manage routines

