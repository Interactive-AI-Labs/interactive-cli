# LLM-Readable Trace & Session Summaries — Design

**Date:** 2026-06-22
**Status:** Approved (design); pending spec review
**Scope:** `interactive-cli` only. No platform/API changes.

## Problem

The CLI's `traces`, `sessions`, and `observations` commands present trace data as
flat metric tables or raw JSON dumps. To understand a single agent turn today, a
reviewer (human or LLM) must chain `traces get` → `observations list --trace-id
--json` → parse raw `langfuse.observation.*` attribute JSON → `observations get`
each interesting span. The flat observation list also discards the
parent/child tree, so the **iteration structure is invisible**. Sessions are
worse: `sessions get` shows a metrics table of turns with **no conversation
content at all** (`SessionTraceSummary` carries no input/output).

The semantic content a reviewer wants already exists in the stored data; it is
just not assembled or rendered. This feature adds an LLM- and human-friendly
**summary rendering** alongside the existing output. Nothing existing is removed
or changed.

### Where the data lives (verified)

The Parlant engine emits a span tree via its own tracer
(`engine.py:100,195,500`), bridged to OTel observations by interactive-agent's
instrumentation. Per turn:

```
process                        ← root span = the turn (one trace)
├─ preparation_iteration_1     ← iteration boundary (parent observation)
│  ├─ match_guidelines         ← output: {matches:[{condition, score, rationale, type, routine_id/policy_id}]}
│  ├─ infer_tool_calls         ← output: {tool_calls:[{tool_name, service_name, arguments}], insights:[...]}
│  └─ execute_tool_calls       ← children named per tool; input=arguments, output=result
├─ preparation_iteration_2 ...
├─ retriever:knowledge_base    ← KB lookups (optional)
├─ message_generation          ← the agent's reply
└─ response_analysis
```

- **Conditions marked true** = `match_guidelines` observation output `matches[]`
  (`condition` + `score`).
- **Tools called** = `execute_tool_calls` child observations (observation
  `name` = tool name, `input` = arguments, `output` = result).
- **Iterations** = group observations by their `preparation_iteration_N` ancestor
  (via `parent_observation_id`).
- **Customer input / agent reply** = trace-level `input` / `output`
  (`langfuse.trace.input` = last customer message; `langfuse.trace.output` =
  JSON array of AI messages).
- **Tools & journeys per turn (session view)** = trace `tags`, which already
  carry `agent:{name}`, `tool:{name}`, and `routine:{journey}`
  (interactive-agent `hooks.py:250-262`). The session view reads these directly,
  avoiding any per-turn observation fetch.

## Goals

1. `iai traces get <id> --summary` renders one turn as a compact narrative:
   per iteration, the conditions met + tools called, then the agent's reply.
2. `iai sessions get <id> --summary` renders the whole conversation as a
   transcript with compact per-turn event tags.
3. Output optimized for LLM review: keep meaningful signal, minimize tokens.
4. Strictly additive: existing commands, flags, columns, and output unchanged.

## Non-Goals

- No `--summary` for `observations get` (a single span is already atomic).
- No new platform endpoint; no change to existing client methods' signatures or
  to any existing output formatter.
- No removal/alteration of `--json`, `--yaml`, `--columns`, `--fields`, etc.

## Design

### Surface

A `--summary` boolean flag added to two commands:

- `cmd/traces.go` → `traces get`
- `cmd/sessions.go` → `sessions get`

`--summary` is mutually exclusive with `--json` and `--yaml`
(`MarkFlagsMutuallyExclusive`). When unset, behavior is byte-for-byte identical
to today. When set, the command fetches the data it needs (see below) and
renders the summary instead of the default detail view.

### Data flow — no client changes

Every existing client method already returns the raw JSON response as a return
value (`GetTrace` → `rawJSON`; `ListObservations` → `rawJSON`; `ListTraces` →
`rawJSON`). The summary layer parses what it needs from that raw JSON, so **no
client method signatures or response structs change**. This also sidesteps the
fact that `TraceInfo` deliberately omits IO fields.

**`traces get --summary`** (2 API calls):
1. `apiClient.GetTrace(..., traceID, "core,io")` → trace raw JSON (input/output/tags/metrics).
2. `apiClient.ListObservations(..., traceID, includeIO=true)` → observations raw JSON (the tree + per-span IO).
3. `summary.TraceSummary(traceRaw, obsRaw)` → `TraceSummaryModel`.
4. `output.PrintTraceSummary(out, model)`.

**`sessions get --summary`** (1 call, + pagination only if >100 turns):
1. `apiClient.ListTraces(..., TraceListOptions{SessionID: id, Fields: "core,io", Order: "asc", Limit: 100})` → traces raw JSON (per-turn input/output + tags).
2. `summary.SessionSummary(tracesRaw...)` → `SessionSummaryModel`.
3. `output.PrintSessionSummary(out, model)`.

> **Implementation contingency to verify first:** confirm the trace-*list*
> endpoint returns `input`/`output` when `fields=core,io`. If it does not, fall
> back to one `GetTrace(..., "core,io")` per turn. Verify before building the
> session renderer.

### Components & boundaries

**`internal/summary/` (new package) — pure transforms, no I/O, no rendering.**
- `TraceSummary(traceRaw, obsRaw json.RawMessage) (*TraceSummaryModel, error)`
- `SessionSummary(traceRaws []json.RawMessage) (*SessionSummaryModel, error)`
- Models:
  - `TraceSummaryModel{ Name, Timestamp, LatencyMs, Cost, Level, Input string, Iterations []Iteration, Reply string, Errors []string }`
  - `Iteration{ Number int, Conditions []Condition, Tools []ToolCall, KBQueries []string }`
  - `Condition{ Text string, Score int }`
  - `ToolCall{ Name, Args, Result string, Errored bool }`
  - `SessionSummaryModel{ ID, Agent string, TurnCount int, Duration, Cost string, Turns []Turn }`
  - `Turn{ Number int, Customer, Agent string, Tools []string, Journeys []string }`
- Parses Langfuse observation JSON; handles the "JSON-encoded string that is
  itself JSON" case (same quirk the existing `prettyJSONUnwrapString` handles).
- Robust to missing/unknown fields: unrecognized spans are skipped, never panic.
  A trace with no recognized observations still renders header + input + reply.

**`internal/output/summary_traces.go` / `summary_sessions.go` (new) — rendering only.**
- `PrintTraceSummary(out io.Writer, m *summary.TraceSummaryModel) error`
- `PrintSessionSummary(out io.Writer, m *summary.SessionSummaryModel) error`
- Reuse existing helpers (`LocalTime`, `formatCost`, `formatLatencyMs`,
  `NewDescribeWriter`, color headers gated on `isTerminal`).

**`cmd/traces.go` / `cmd/sessions.go` — wiring only.**
- Add `--summary` flag var + registration; branch in `RunE` before the existing
  JSON/YAML/detail branches.

### Output formats

**Trace summary** (compact narrative; see mockup):

```
Turn — driveaway-agent · 2026-06-22 14:32:01 · 4.2s · $0.012 · 2 iterations

Customer: I want to rent a car for next weekend

Iteration 1
  Conditions met:
    ✓ Customer asks to rent a vehicle (9)
    ✓ No booking in progress (7)
  Tools called:
    → check_availability(dates="next weekend") → 3 cars found
Iteration 2
  Conditions met:
    ✓ Pickup location not yet provided (8)
  (no tools called)

Agent: Great! We have 3 cars available for next weekend...
```

If `level == ERROR`, append ` · ERROR` to the header and list errored
observations (name + status message) in an `Errors:` block.

**Session summary** (transcript + event tags; see mockup):

```
Session s_abc · driveaway-agent · 6 turns · 3m12s · $0.08

Turn 1  Customer: I want to rent a car next weekend
        [tools: check_availability]
        Agent: Great! We have 3 cars available...
Turn 2  Customer: The SUV please
        Agent: Got it. Where will you pick up?
Turn 3  Customer: Downtown
        [tools: create_booking] [journey: Rental→done]
        Agent: Booked! Confirmation #1234
```

Event tags shown only when present. `[tools: ...]` from `tool:*` trace tags;
`[journey: ...]` from `routine:*` trace tags. Agent label/name from `agent:*`
tag (fallback: trace name prefix).

### Token discipline

- Long values (tool args, tool results, customer/agent messages) truncated to a
  default cap of **500 characters** with a trailing `… (truncated)`.
- Guideline **rationale is omitted** by default (condition + score only); still
  available via existing `--json`.
- Only recognized spans are rendered; internal/structural noise dropped.

### Errors & edge cases

- Trace/observations fetch error → surface the underlying client error (same as
  existing commands).
- Trace with zero observations → header + input + reply only (no iteration
  blocks).
- Session with zero traces → `No turns found.` (mirrors existing empty-state
  messages).
- Tool result that is an error (Parlant sentinel or built-in `ok:false`) →
  mark the `ToolCall.Errored` and render `→ <name>(...) → ERROR: <message>`.
- Non-JSON or unexpected IO payloads → render the raw string (truncated), do not
  fail the whole summary.

## Testing

- `internal/summary/*_test.go`: table-driven unit tests over hand-built fixture
  JSON matching the documented Langfuse attribute shapes — multi-iteration trace,
  trace with tool errors, trace with no observations, JSON-encoded-string IO,
  session with/without event tags, empty session.
- `internal/output/summary_*_test.go`: golden-style assertions on rendered text
  for representative models (TTY color off).
- Follow existing test conventions in `internal/output/*_test.go` and
  `internal/inputs/*_test.go`.

## Documentation

- `docs/` command reference is auto-generated (cobra). Regenerate so
  `iai_traces_get.md` and `iai_sessions_get.md` reflect `--summary`
  (via the existing docs-gen make target / script).
- Add `--summary` usage to each command's cobra `Example` block.

## Rollout

Single PR on `feature/llm-readable-trace-summaries`. Additive only; no migration,
no breaking change.
