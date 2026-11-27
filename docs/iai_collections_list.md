## iai collections list

List collections in a database, or the one an agent searches

### Synopsis

List every collection in a database (-d), or only the collection an agent's
manifest searches (--agent, which defaults to the selected agent). With both, the
list is empty unless the agent's collection lives in that database.

```
iai collections list [flags]
```

### Examples

```
  iai collections list -d my-db
  iai collections list -d my-db --json
  iai collections list --agent my-agent
```

### Options

```
      --agent string          Filter by agent name; defaults to the selected agent
  -d, --database string       Database to list; required without --agent or a selected agent
  -h, --help                  help for list
      --json                  Output raw API response as JSON
  -o, --organization string   Organization name
  -p, --project string        Project name
      --yaml                  Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai collections](iai_collections.md)	 - Knowledge bases (searchable tables of chunks) inside a pgvector database

