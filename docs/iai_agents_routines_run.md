## iai agents routines run

Launch an existing autonomous routine

### Synopsis

Launch an existing autonomous routine and return its receipt without waiting
for completion. Inspect tool activity and output with 'iai agents events list'.
Requests are not automatically retried.

```
iai agents routines run <agent_name> <routine_id> [flags]
```

### Examples

```
  iai agents routines run my-agent billing/refund --input-file input.json --callback-url https://trusted.example/results
  iai agents routines run my-agent billing/refund --input-file input.json --callback-url https://trusted.example/results --session-id sess_123
  cat input.json | iai agents routines run my-agent billing/refund --input-file - --callback-url https://trusted.example/results --idempotency-key refund-order-123 --json
```

### Options

```
      --callback-url string      Trusted HTTP(S) URL for the authoritative completion result; session events may omit final failures. Required by the agent, allowed by the routine, and receives the agent's bearer credential
  -h, --help                     help for run
      --idempotency-key string   Key to prevent duplicate execution; duplicates print the existing run and exit with an error; deduplication does not survive agent restarts
      --input-file string        Routine input as a JSON object, from a file or stdin (-); not a routine definition
      --json                     Output raw receipt as JSON; session notices and follow instructions go to stderr
  -o, --organization string      Organization name
  -p, --project string           Project name
      --session-id string        Use this existing session; when omitted, create a retained session without a greeting. Created sessions survive failed or duplicate triggers; their IDs go to stderr
      --yaml                     Output raw receipt as YAML; session notices and follow instructions go to stderr
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents routines](iai_agents_routines.md)	 - Execute autonomous routines on a deployed agent

