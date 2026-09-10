## iai agents replay

Test an agent by replaying recorded scenarios and report the verdict

### Synopsis

Test a deployed agent by replaying recorded scenarios and report the verdict.

A scenario is a dataset item: a recorded conversation's customer messages,
context and tool results, plus what the run must satisfy. The platform re-runs
it against the agent with live reasoning, answers tool calls from the recorded
fixtures, checks the expectations and writes PASS/FAIL scores. Replay a whole
dataset (--dataset, optionally narrowed with --scenarios), a local scenario
file (--file), or re-attach to a run in flight (--run-id).

A replay writes a synthetic customer and session to the agent and uses your
project's quota: prefer a non-production agent. Progress goes to stderr and
only the verdict to stdout. Exit code is 0 for any verdict, 1 when the replay
could not run; gate in CI with --json and jq -e '.status == "passed"'.

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
```

### Options

```
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
      --timeout duration        Give up waiting for the verdict after this long (default 30m0s)
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

