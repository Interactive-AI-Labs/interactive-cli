## iai agents create

Create an agent in a project

### Synopsis

Create an agent in a specific project.

The --file flag takes a YAML file matching the agent_config schema. Pass the
agent name as the positional argument and id/version/env/secrets/endpoint/schedule
via flags; do not include them inside the file.

The config schema depends on the agent version. Run
'iai agents compatibility-matrix' to find which schema version applies, then
'iai agents schema --schema-version <schema>' to see the expected fields.

Routines and policies referenced in the config must already exist in the project
and should be validated against the matching schema version (see --schema-version
on their create/update commands).

```
iai agents create <agent_name> [flags]
```

### Examples

```
  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml
  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml --endpoint
  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml --secret api-keys --env LOG_LEVEL=info
```

### Options

```
      --endpoint                   Expose the agent at <agent-name>-<project-hash>.interactive.ai
      --env stringArray            Environment variable (NAME=VALUE); can be repeated
      --file string                Path to YAML file matching the agent_config schema (run 'iai agents schema' to see it)
  -h, --help                       help for create
      --id string                  Agent type from the marketplace (e.g. interactive-agent)
      --mcp stringArray            Attach an MCP by name (see 'iai mcps list'); can be repeated
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

