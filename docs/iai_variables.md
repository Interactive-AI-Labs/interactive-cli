## iai variables

Manage variables

### Synopsis

Manage variables in InteractiveAI projects.

Variables are persistent data fields that survive across conversation sessions
(JSON format).


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
  -h, --help   help for variables
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
* [iai variables create](iai_variables_create.md)	 - Create a variable
* [iai variables delete](iai_variables_delete.md)	 - Delete a variable
* [iai variables get](iai_variables_get.md)	 - Get details of a variable
* [iai variables list](iai_variables_list.md)	 - List variables in a project
* [iai variables update](iai_variables_update.md)	 - Update a variable (creates a new version)

