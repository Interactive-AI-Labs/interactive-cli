## iai agents update

Update an agent in a project

### Synopsis

Update an agent in a specific project.

Only the flags you pass are applied; everything else is left at its current
value.

--file takes a YAML file matching the agent_config schema and replaces the
entire agent config in full when provided (no per-field merge). The config
schema depends on the agent version — run 'iai agents compatibility-matrix'
to find which schema version applies, then 'iai agents schema --schema-version <schema>'
to see the expected fields.

When upgrading to a new agent version with a different schema, update your
routines and policies first using --schema-version on their create/update
commands, then update the agent with the new config and version.

Lists (--env, --secret) replace the entire current list when provided — pass
every value you want to keep.

For schedules, passing --schedule-uptime auto-clears any existing downtime,
and --schedule-downtime auto-clears any existing uptime. Pass --schedule-timezone
alongside either to change the timezone.

Use --clear-env, --clear-secret, --clear-schedule, or --clear-stack-id to
remove those configurations entirely.

--detach-mcp removes an mcp reference (bare or resolved) by name; combine
with --mcp in the same command to swap one for another. Detach an mcp before
deleting it — 'iai mcps delete' blocks by default while an agent still
references it.

```
iai agents update <agent_name> [flags]
```

### Examples

```
  iai agents update chat-agent --version 0.0.3
  iai agents update chat-agent --file agent-config.yaml
  iai agents update chat-agent --endpoint=false
  iai agents update chat-agent --schedule-uptime "Mon-Fri 07:30-20:30" --schedule-timezone Europe/Berlin
  iai agents update chat-agent --clear-schedule
  iai agents update chat-agent --stack-id my-stack
  iai agents update chat-agent --clear-stack-id
  iai agents update chat-agent --mcp github --mcp stripe
  iai agents update chat-agent --detach-mcp stripe
```

### Options

```
      --clear-env                  Remove all environment variables from the agent
      --clear-schedule             Remove the schedule configuration from the agent
      --clear-secret               Remove all secret references from the agent
      --clear-stack-id             Remove the agent from its stack
      --detach-mcp stringArray     Detach an MCP by name; can be repeated. Without --file, removes from the agent's current mcps (applied before --mcp)
      --endpoint                   Expose the agent at <agent-name>-<project-hash>.interactive.ai
      --env stringArray            Environment variable (NAME=VALUE); can be repeated
      --file string                Path to YAML file matching the agent_config schema (run 'iai agents schema' to see it)
  -h, --help                       help for update
      --id string                  Agent type from the marketplace (e.g. interactive-agent)
      --mcp stringArray            Attach an MCP by name (see 'iai mcps list'); can be repeated. Without --file, appends to the agent's current mcps
  -o, --organization string        Organization name
  -p, --project string             Project name
      --schedule-downtime string   When the agent should be scaled down (mutually exclusive with --schedule-uptime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Example: 'Sat-Sun 00:00-24:00'
      --schedule-timezone string   IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); required with --schedule-uptime or --schedule-downtime
      --schedule-uptime string     When the agent should be running (mutually exclusive with --schedule-downtime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Example: 'Mon-Fri 07:30-20:30'
      --secret stringArray         Secret to inject as environment variables; can be repeated
      --stack-id string            Stack ID to assign the agent to
      --version string             Agent image version to deploy (e.g. 0.0.1)
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

