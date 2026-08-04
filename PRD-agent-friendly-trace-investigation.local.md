# PRD — Agent-Friendly Trace Investigation for the `iai` CLI

**Status:** Draft v2 (local working doc — not committed, not shared)
**Author:** Oliver Ryall (with Cursor / Claude)
**Date:** 2026-08-03
**Motivating exercises:** Live Betsson `customer-support` investigation, last 6 hours (2026-08-03 ~03:00–09:00 UTC); bulk review of the last 200 Betsson traces (2026-08-03); live Alunafi `company-onboarding` error hunt, last 10 hours (2026-08-03).

---

## Problem Statement

FDEs (and coding agents acting for them) regularly ask an LLM to “look at the last few hours/days of traces for client X and find errors or weird patterns.” That workflow today is token-expensive and often misleading — not because the HTTP APIs are slow (every call is sub-second), and not everywhere: credit where due, **the labeled-error hunt already works well**. In the Alunafi exercise, one `traces list` with a `level` column spotted 9 error traces instantly, and one cross-trace `observations list --level ERROR --columns trace_id,name,level,status_message` returned the complete error inventory for a 10-hour window in a single call (~8 CLI invocations end to end). The defects are elsewhere — exactly two:

1. **The compact path is silently broken.** `traces get --summary` — the CLI's purpose-built LLM-readable view — reports **0 iterations, no tools, no journey, no conditions** on current production traces. The summary walker (`internal/summary/trace.go`) matches legacy engine span names (`preparation_iteration_N`, `match_guidelines`, `execute_tool_calls`, `next-step`) while the platform now emits display names (`Iteration: N`, `Evaluate: Tools`, `Execute: Tools`, `Next step: …`, `Search: Policies & Routines`). Verified live twice: a Betsson KYC trace with 4 iterations and 7 tool calls summarized to just input + reply; in Alunafi, error traces with 555+ observations showed "0 iterations" — hiding the single most important fact (the trace ran ~11 concurrent sub-flows, one per company member, with tool failures inside), which had to be reconstructed by dumping observation names through `sort | uniq -c`. (Implementation follow-up: start times proved these are parallel sub-runs, not restarts — the fixed summary orders iterations chronologically by `start_time`, which renders both cases truthfully.) `traces diff` builds on the same model, so it compares empty iteration lists. Session `--summary` only escapes because it derives tools/journeys from trace *tags*, not observations. Related failure: a real Betsson KYC issue sat at `WARNING` with `status_message: no routine matched` — invisible in the summary, which collects ERROR only. Agents get a false “all clear,” lose trust, and fall back to raw dumps. The Betsson chat exercise added the sharpest consequence: a turn whose reply *claims* “I retrieved your withdrawal details… your €1000 withdrawal is registered” while its only tool-arg generation errored with `JSONDecodeError` — with tools/iterations missing from the summary, an agent **cannot tell whether the tool eventually succeeded on retry or the reply is fabricated**, which is precisely the judgment a trace review exists to make.

2. **Unlabeled (semantic) review has no bulk path.** When the issue is not marked ERROR — a wrong reply, a bad routine decision, a WARNING-level “no routine matched” — the agent must read what each turn actually *did*, and depth is per-trace only. Reviewing a window means list → then `traces get`/`observations list` per ID. Agents batch this in a shell loop rather than 200 separate tool calls, but batching doesn't fix the real costs: the compact view per trace is broken (defect 1), and the working alternatives flood the context window — `traces list --fields core,io,metrics --json` is **~4.7 MB per 100 traces**; per-trace raw fallback (`traces get --json` + `observations list --include-io`) is ~34 KB–1.4 MB per trace. In practice the agent samples 5–10 traces and extrapolates — coverage is exactly what suffers on “find weird patterns” asks.

The CLI’s job is still to be a **thin, reliable client**: parse → call platform → print. It should not invent “bug finding brains.” It *should* make the existing observe verbs cheap and trustworthy for agents that do the finding.

## Solution

Two changes, nothing else. After them, a typical agent investigation is **filter once, read compact once**.

1. **Fix `traces get --summary`** so it is correct for current production observation names (keeping legacy names working), and include WARNING status lines — not only ERROR — so “no routine matched” class issues surface in the compact path. This transitively repairs `traces diff`.
2. **Add `--summary` to `traces list`** (behavioral flag on the existing verb — no new command), so one invocation returns compact summaries for a filtered page of traces. The CLI does the per-trace fan-out internally with a bounded worker pool. This exists for the **semantic-review** case — labeled-error hunting already has an efficient one-call path via cross-trace `observations list --level ERROR` and should keep using it.

No new CLI verbs. No platform API changes. No automated “find the bug” logic. No convenience flags beyond these two changes — everything else that came up (relative `--since` durations, observation default-column changes, capped `--all` crawling) is deliberately deferred; see Out of Scope.

### Target agent recipe after this ships

```bash
# Labeled errors: already one call today — unchanged
# (--tags agent:<name> scopes to one agent when a project deploys several)
iai traces list --from-timestamp 2026-08-01T00:00:00Z --tags agent:agent-chat \
  --columns id,name,timestamp,latency,level,cost
iai observations list --from-timestamp 2026-08-01T00:00:00Z --level ERROR \
  --columns trace_id,name,level,status_message

# Semantic review ("any issues / weird patterns"): compact read of the window in one shot
iai traces list --from-timestamp 2026-08-03T00:00:00Z --limit 100 --summary --json

# When needed, one turn deeper (now trustworthy)
iai traces get <id> --summary --json
```

## User Stories

1. As an FDE, I want `iai traces get <id> --summary` to show iterations, tools, journey steps, and errors for current production agents, so that I can understand a turn without dumping full observation IO.
2. As a coding agent, I want summary JSON to include iterations/tools/errors when they exist in observations, so that I do not fall back to multi‑MB `--include-io` payloads.
3. As an FDE, I want legacy observation names to keep summarizing correctly, so that older traces and engines do not regress.
4. As a coding agent, I want WARNING observations with a `status_message` to appear in the summary (not only ERROR), so that cases like “no routine matched” are visible in the compact path.
5. As an FDE, I want `iai traces list --summary` to reuse existing list filters (`--has-error`, `--level`, `--tags`, `--min-latency`, `--search`, time range, etc.), so that I do not learn a second query language.
6. As a coding agent, I want `traces list --summary --json` to return an array of summary models (not raw list envelopes), so that I can parse one stable shape — and reviewing 200 traces means 2 compact invocations instead of a shell loop of 200 gets whose output floods my context.
7. As an FDE, I want `traces list --summary` without `--json` to print the same human summary rendering used by `traces get --summary`, so that interactive use stays consistent.
8. As a coding agent, I want list `--summary` bounded by `--limit` / `--page` (max 100 as today), so that I cannot accidentally request unbounded work.
9. As a coding agent, I want partial failures during list-summary (one trace get fails) to be reported per-item without aborting the whole batch, so that I still get the rest of the window.
10. As an FDE, I want existing `--json` raw behavior preserved when `--summary` is off, so that scripts depending on API envelopes do not break.
11. As an FDE, I want no new verb such as `traces scan` / `traces investigate`, so that the CLI verb set stays consistent with `list`/`get`.

## Implementation Decisions

### Ethos constraints (non-negotiable)

- Thin client: parse flags → call existing platform endpoints → project/print. No bug-clustering “brain,” no new observability product surface.
- Flags on existing verbs only (`list`/`get`).
- Keep `RunE` thin; matching and batch orchestration live in `internal/`.
- Reuse the existing summary model and printers; do not invent a second summary schema.
- `--json` / `--yaml` with `--summary` mean **structured summary models** (as `traces get --summary --json` already does), not raw API envelopes. Without `--summary`, today’s raw behavior is unchanged.

### Change 1: fix the summary walker (`internal/summary`)

- Accept **both** legacy span names and current display names observed in production:
  - Iterations: `preparation_iteration_N` **and** `Iteration: N`
  - Guideline/context matching: `match_guidelines` **and** the current spans carrying the same match payload (live Betsson: `Evaluate: Context` / `Evaluate: Routine steps`) — match by known display names plus payload shape where a single string would be brittle
  - Tool execution: `execute_tool_calls` **and** `Execute: Tools` (tools remain child TOOL spans)
  - Next-step decisions: `next-step` **and** `Next step: ` prefix
  - Knowledge-base spans: keep existing retriever matching; extend to `Search: …` display names as observed
- Collect status lines for **ERROR and WARNING** into the summary’s errors list (WARNING was the only real issue in the motivating window).
- Extend existing summary model fields only; no agent-specific narrative text.
- `traces diff` needs no change of its own — it is fixed by this.

### Change 2: `traces list --summary`

- Add `--summary` to `traces list`, composed with all existing filters.
- Behavior: run the normal list query → for each returned trace, build the same summary used by `traces get --summary` (trace fields + observations with IO) → print in list order.
- Parallelize per-trace fetches with a small bounded worker pool (e.g. 8) in an `internal/` helper; output order stays stable.
- Scope of work: current page only (`--page` / `--limit`, max 100 as today). No `--all`, no multi-page crawling.
- Flag rules:
  - `--summary` mutually exclusive with `--columns` (summary is a view, same spirit as get).
  - `--summary` composes with `--json` / `--yaml` like get.
  - Per-trace failure → structured per-item error entry (or stderr warning + skip); the command fails only if the list itself failed.
- Update `docs/iai_traces_list.md` and command examples for the new flag (required by repo conventions; nothing beyond that).

### Explicitly not chosen (ethos / scope)

- **No `--since <duration>` sugar** — agents compute RFC3339 timestamps fine; it is not what makes the workflow expensive. Revisit only if it keeps hurting after this ships.
- **No observation default-column changes** (`level`, `status_message`) — the WARNING visibility gap is closed at the summary layer instead, which is where agents read.
- **No capped `--all` pagination on list-summary** — one page of 100 is enough for “review the last N traces”; two calls covers 200.
- **No new platform bulk-summary or error-aggregate API.**
- **No NDJSON format flag** — existing structured `--json` printers.
- **No tag/agent filter on cross-trace observation search** — scoping error observations to one agent means grepping results against trace IDs from a tag-filtered `traces list`; acceptable client-side join at observed scales, and adding the filter is a platform API question, not CLI projection.
- **No automatic anomaly detection** — product logic, not thin-client projection.

## Testing Decisions

Good tests assert **external behavior of seams**, not internal regex variables or private helpers.

### Summary package (highest value)

- Table-driven tests with **live-shaped** observation names/payloads:
  - Chat turn with `Iteration: N`, `Execute: Tools` + TOOL children, `Next step: …`
  - KYC turn with many iterations and tool sequence (mirror the verified live case: 4 iterations, 7 tool calls)
  - WARNING with `status_message` appears in summary
  - Legacy `preparation_iteration_N` / `match_guidelines` / `execute_tool_calls` still pass (regression)
- Assert on the summary model (JSON structure), following existing summary tests.

### List `--summary` orchestration

- Fake/stub API client returning a small list + per-trace detail/observations.
- Assert: N summaries produced; order preserved; `--json` is structured summaries; one failing get does not drop siblings; `--columns` + `--summary` rejected; raw `--json` unchanged when `--summary` off.
- Prefer testing the internal helper over a giant cobra integration test, matching nearby patterns.

## Out of Scope

- Relative-duration time flags (`--since 6h`) on any command
- Observation list default-column changes
- Platform endpoints for bulk summaries, error aggregates, or higher list limits
- New CLI verbs (`scan`, `investigate`, `analyze`, …)
- Changing auth, org/project selection, or multi-project fan-out
- Automatic bug/pattern detection or LLM calls inside the CLI
- Redesigning session summary (already good enough for multi-turn chat; note it reads trace tags, not observations, so it is unaffected by Change 1)
- Changing raw `--json` envelope shape for non-summary list/get
- KYC/chat product fixes themselves (e.g. “no routine matched” for GREEN onHold) — this PRD only makes them discoverable cheaply
- Performance work on the observability backend

## Further Notes

### Motivating command logs (abridged)

**Betsson `customer-support`, 6h window (semantic review — the painful case):**

1. Discover/select org + project + agents
2. Compute `now-6h` in the shell
3. `metrics list --daily`, `traces list`, `--has-error`, `--level ERROR|WARNING`
4. `observations list --level WARNING` → `observations get` for `status_message`
5. `traces list --min-latency` / `--min-cost` / `--tags` / `--search`
6. Several `traces get --summary` (returned 0 iterations) → fall back to `observations list`
7. `sessions get --summary` once `session_id` known

API latency per call was sub‑second; the cost was **broken summary trust and huge IO dumps**.

**Alunafi `company-onboarding`, 10h window (labeled-error hunt — mostly fine today):**

1. `organizations list` + `projects list -o Alunafi`
2. Compute `now-10h` in the shell
3. `traces list --from-timestamp … --columns id,name,timestamp,latency,level,observation_count,cost` → 30 traces, 9 ERROR, and the anomaly visible at a glance (error traces: 518–655 observations, 150–290 s, $0.34–0.54 vs 4 observations / ~$0.003 healthy)
4. `observations list --level ERROR --from-timestamp … --columns trace_id,name,level,status_message` → **complete error inventory in one call** (`step_not_registered` in all 9; `output_validation_failed` schema mismatch in 5; one 429; one `tin_check` with empty TIN)
5. `traces get --summary` on an error trace → Errors list useful, but "0 iterations" on a 555-observation trace
6. Three extra calls reconstructing the retry-loop structure via observation-name dumps + `sort | uniq -c` — needed only because the summary hid it

~8 invocations total. The efficient steps (3–4) need no change. Steps 5–6 are defect 1 in action: the summary should have said "74 iteration spans, flow restarted 11×, jira tools called 18×" directly.

**Betsson `customer-support` chat agent, 3-day window (labeled-error hunt on a busier project):**

1. `traces list --tags agent:agent-chat --from-timestamp …` (all / `--level ERROR` / `--level WARNING`) → 615 traces, 8 ERROR, 0 WARNING. `--tags agent:<name>` cleanly scoped one agent out of several in the project; the list footer (“Page 1 of 7 (615 total items)”) made truncation visible.
2. Cross-trace `observations list --level ERROR --columns trace_id,name,level,status_message` → full inventory in one call, **but** grepped client-side down to the 8 chat trace IDs because observation search has no tag/agent filter — fine at this scale, a join the agent must remember to do.
3. `traces get --summary` ×3 for user impact → Errors + reply useful: one turn visibly sent the customer four near-duplicate “Je viens de vérifier…” messages while `get_bonus_eligibility` was retried 4× missing `party_id`; another replied with a confident withdrawal confirmation after its only tool-arg generation threw `JSONDecodeError`. With iterations/tools hidden ("0 iterations" again), it was impossible to tell from the summary whether that tool later succeeded — recovered-vs-fabricated is unanswerable through the compact path today.

5 invocations total. Same verdict: the filter/list layer is healthy; every remaining blind spot in all three exercises is the summary.

### Measured costs (live, 2026-08-03)

Per turn:

| View | Approx size |
|---|---|
| `traces get --summary --json` | ~0.7 KB (incomplete today) |
| `observations list` table (no IO) | ~7 KB |
| `traces get --json` | ~4–66 KB |
| `observations list --include-io --json` | ~29 KB–1.4 MB |

At bulk scale (the “review last 200 traces for issues” semantic ask):

| Approach today | Shape | Size / time |
|---|---|---|
| `traces list` table ×2 pages | 2 calls | ~40 KB — metadata only, no content to judge |
| `traces list --fields core,io,metrics --json` ×2 | 2 calls | ~9.4 MB — context-window trap |
| `traces get --summary` per trace | shell loop of 200 gets (typically batched in one Bash call) | ~0.65 s each ⇒ 2+ min serial fetching, and today each summary is incomplete anyway |

After this ships: 2 invocations (`traces list --summary --json`, two pages), ~15–20 s inside the CLI (bounded parallel fan-out), ~200–400 KB of trustworthy compact content covering all 200 traces.

### Success bar

Two live-validated checks:

- **Betsson-style semantic ask** (“inspect last 6h for issues”): answerable primarily with one filtered `traces list … --summary --json`, without `--include-io` unless the agent is debugging a specific span payload.
- **Alunafi-style error trace**: `traces get --summary` on a 555-observation trace must directly show the iterations, parallel sub-runs, and tool calls that are actually in it (e.g. “74 iterations across ~11 concurrent sub-flows, jira tools called 18×”) — no `sort | uniq -c` reconstruction, no false “0 iterations”. ✅ Verified live during implementation (2026-08-03): header reads “74 iterations”, ordered chronologically.

The already-good path stays untouched: cross-trace `observations list --level ERROR --columns …,status_message` remains the one-call answer for labeled errors.
