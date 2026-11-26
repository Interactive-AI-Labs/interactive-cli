## iai agents traces

List an agent's traces

### Synopsis

List an agent's traces.

Without agent_name, the selected agent is used (see 'iai agents select').

```
iai agents traces [agent_name] [flags]
```

### Examples

```
  iai agents traces my-agent
  iai agents traces my-agent --limit 20 --json
```

### Options

```
      --columns strings           Columns to display for table output only (comma-separated, default: id,name,timestamp,latency,cost,tags). Cannot be used with --json or --yaml.
                                  Available: id,name,timestamp,user_id,session_id,release,version,environment,public,latency,cost,tags,observation_count,input_tokens,output_tokens,total_tokens,level
      --environment stringArray   Filter by environment (repeatable)
      --fields string             Field groups to include: core, io, metrics (comma-separated) (default "core,metrics")
      --from-timestamp string     Filter traces from this timestamp (ISO 8601, default: 7 days ago)
      --has-error                 Filter traces with errors
  -h, --help                      help for traces
      --json                      Output raw API response as JSON
      --level string              Filter by aggregated level (DEBUG|DEFAULT|WARNING|ERROR)
      --limit int                 Items per page (max 100)
      --max-cost float            Maximum total cost filter
      --max-latency float         Maximum latency filter (seconds)
      --max-tokens int            Maximum total tokens filter
      --min-cost float            Minimum total cost filter
      --min-latency float         Minimum latency filter (seconds)
      --min-tokens int            Minimum total tokens filter
      --model string              Filter by model name
      --name string               Filter by trace name
      --order string              Sort direction (asc|desc) (default "desc")
      --order-by string           Order by field (timestamp|latency|cost|name)
  -o, --organization string       Organization name that owns the project
      --page int                  Page number (starts at 1) (default 1)
  -p, --project string            Project name
      --release string            Filter by release
      --search string             Search in trace name (max 200 characters)
      --session-id string         Filter by session ID
      --summary                   Show conditions, decisions, tools, and results for each listed turn
      --tags stringArray          Filter by tags (repeatable)
      --to-timestamp string       Filter traces to this timestamp (ISO 8601)
      --user-id string            Filter by user ID
      --version string            Filter by version
      --yaml                      Output raw API response as YAML
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools

