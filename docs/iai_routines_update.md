## iai routines update

Update a routine (creates a new version)

### Synopsis

Update a routine by creating a new version with updated content.

This creates a new version of the routine using the content from the provided file.
The previous versions are preserved and can still be accessed by version number.


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
      --file string           Path to the file containing the updated prompt content
  -h, --help                  help for update
      --labels strings        Labels for the new prompt version (comma-separated)
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --skip-schema           Skip schema validation (allows draft/WIP content)
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

