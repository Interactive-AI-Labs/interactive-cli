## iai variables create

Create a variable

### Synopsis

Create a new variable definition in an InteractiveAI project.

Content is provided via a JSON file using the --file flag and must follow the
variable schema below.


### Schema

```json
{
  "variables": [
    {
      "name": "<string>",
      "type": "<boolean|string|number|array|object>",
      "persistence": "<session|customer|global>",
      "default_value": "<any>"
    }
  ]
}
```

> `name` and `type` are required. `persistence` defaults to `"session"`. `default_value` is optional.

### Example

```json
{
  "variables": [
    {"name": "user_name", "type": "string"},
    {"name": "is_authenticated", "type": "boolean", "default_value": false},
    {"name": "preferences", "type": "object", "persistence": "customer"}
  ]
}
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

* [iai variables](iai_variables.md)	 - Manage variables

