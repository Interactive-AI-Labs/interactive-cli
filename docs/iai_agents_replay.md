## iai agents replay

Replay recorded scenarios against an agent and report the verdict

### Synopsis

Replay recorded scenarios against a deployed agent and report the verdict.

A scenario is a recorded conversation stored as a dataset item: customer
messages, context variables, the tool results the agent saw, and optionally
what the run must satisfy. Replaying re-runs those inputs with live reasoning,
answers tool calls from the recorded fixtures, then checks the expectations
and writes PASS/FAIL scores to the platform.

The command talks to the agent itself over its public hostname, not to the
platform. Any agent running agent-server 0.15.0 or later serves the replay
routes; there is nothing to enable. Three modes:

  --dataset D            replay every scenario in the dataset, or only the
                         names given with --scenarios
  --file PATH            replay a local scenario file (YAML or JSON) inline
  --run-id ID            re-attach to a run already started on the agent

Iterations run in parallel on the agent, up to --concurrency across the whole
run. A replay writes a synthetic customer, a session, and variable values to
the target agent and spends router quota: prefer a non-production agent.

The agent's API key is taken from --agent-api-key, then INTERACTIVE_AGENT_API_KEY,
then resolved from the agent's own configuration by reading the project
secret it mounts. That last step needs secret-read permission, so CI should
set INTERACTIVE_AGENT_API_KEY instead.

Exit code is 0 when the run finished with a verdict, passed or failed, and 1
when the replay could not run. The verdict itself is in the output; gate on
it in CI with --json and jq -e '.status == "passed"'.

```
iai agents replay <agent_name> [flags]
```

### Examples

```
  iai agents replay agent-chat-dev --dataset replay-chat --scenarios account-lock
  iai agents replay agent-chat-dev --dataset replay-chat --scenarios account-lock --scenarios bonus-misrouted --repeat 3
  iai agents replay agent-chat-dev --dataset replay-chat --repeat 3 --concurrency 16
  iai agents replay agent-chat-dev --file ./account-lock.yaml
  iai agents replay agent-chat-dev --dataset replay-chat --json > run.json
  iai agents replay agent-chat-dev --run-id 9d0c44e1aa52
  iai agents replay my-local-agent --file ./x.yaml --agent-url http://127.0.0.1:8080 --agent-api-key "$AGENT_API_KEY"
```

### Options

```
      --agent-api-key string    Bearer for the agent (else INTERACTIVE_AGENT_API_KEY, else resolved from the agent's secrets)
      --agent-url string        Agent base URL, overriding the public hostname (e.g. http://127.0.0.1:8080)
      --concurrency int         In-flight iterations across the whole run (1-32) (default 8)
      --dataset string          Dataset holding the scenarios (required unless --file or --run-id)
      --file string             Local scenario file (YAML or JSON), posted inline
  -h, --help                    help for replay
      --json                    Print the final run payload exactly as the agent returned it; progress still goes to stderr
  -o, --organization string     Organization name
  -p, --project string          Project name
      --repeat int              Iterations per scenario (1-20) (default 1)
      --run-id string           Re-attach to a run already started on the agent; no new run is started
      --scenarios stringArray   Only this scenario name from --dataset; repeat the flag once per scenario, run in this order
      --timeout duration        Give up polling after this long (default 30m0s)
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

