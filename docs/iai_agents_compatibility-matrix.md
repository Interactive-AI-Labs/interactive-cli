## iai agents compatibility-matrix

Show agent version to schema version compatibility

### Synopsis

Display the compatibility matrix between agent versions and schema versions.

Each agent version requires a specific config schema. Use this command to find
the schema version for your target agent version, then run
'iai agents schema --schema-version <schema>' to see the expected config fields.

Prompt types (routines, policies, etc.) also support versioned schemas — use
--schema-version on their create/update commands to validate against the
matching version.

By default, output is a formatted table. Use --json for machine-readable output.

```
iai agents compatibility-matrix [flags]
```

### Examples

```
  iai agents compatibility-matrix
  iai agents compatibility-matrix --json
```

### Options

```
  -h, --help   help for compatibility-matrix
      --json   Output raw JSON instead of a formatted table
      --yaml   Output structured YAML
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

