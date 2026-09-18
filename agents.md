# Agents guidelines for `interactive-cli`

This repo is a **Go CLI** using **Cobra**. It is a **thin client** for the Interactive AI platform:

- Authenticates the user.
- Collects flags / args / env.
- Calls HTTP APIs.
- Prints results (tables, text, errors).

The platform is the source of truth. The CLI should not implement “agent brains” or complex business logic.

---

## 1. Where things live

- `cmd/`  
  - Cobra commands and CLI wiring only.  
  - Responsibilities:
    - Command usage/flags.
    - Parsing args/env.
    - User-facing messages.
    - Delegating to `internal/`.

- `internal/`  
  - Shared helpers for HTTP, storage, output formatting, config, lookups, etc.  
  - Responsibilities:
    - Build URLs and requests.
    - Handle responses and errors.
    - Render tables / text.
    - Load/store session and config.

**Rule:** Keep `RunE` bodies small. Put real work in `internal/` functions and call them from `cmd/`.

---

## 2. How to work in this repo

1. **Locate existing code**
   - Find the closest command in `cmd/`.
   - Look for reusable helpers in `internal/`.
   - Prefer extending existing patterns over new ones.

2. **Plan a small change**
   - One feature or bugfix per change.
   - Avoid broad refactors unless explicitly requested.
   - Preserve public behavior unless the task says otherwise.

3. **Implement**
   - New CLI behavior → new/updated command in `cmd/`.
   - HTTP / formatting / config logic → helpers in `internal/`.
   - Keep functions short and focused; avoid deep abstraction layers.

4. **Validate**
   - Ensure it compiles.
   - Conceptually run:
     - `go test ./...` or at least affected packages.
   - Keep logic simple and predictable.

---

## 3. Coding rules (concise)

### Structure and logic

- **No complex logic in `Run`/`RunE`.**
  - Do argument/flag handling and basic validation there.
  - Call small helpers in `internal/` for the rest.

- **Keep related code together.**
  - HTTP client + types in the same package/file when practical.
  - A command should either call `internal` directly or via a small local helper.

### Comments and dead code

- **No commented-out code.** Delete unused code; rely on git history.
- **Comments only when needed.**
  - Explain *why*, not *what*.
  - Document tricky behavior or external API quirks.
- **Keep them short.** One or two lines. A comment longer than the code it
  describes is documentation in the wrong place: cut it to the one fact the
  code cannot state, or move it to the command's help text.
- **Do not restate the code.** A comment that paraphrases the line below it is
  noise, and it goes stale on the next edit.
- **Write for someone using the code, not reviewing the diff.** Reasoning that
  justifies a decision belongs in the commit message or the PR.
- Remove unused functions, types, and imports. No `_ = someName` hacks.

### Dependencies and init

- Use versions pinned in `go.mod`.
- Add new dependencies only when the stdlib + existing deps are insufficient.
- Keep `init()` light:
  - No goroutines, network calls, or heavy work at import time.
  - Prefer explicit wiring from commands.

### Environment and config

- Read env/config in a small number of central places (e.g., near `rootCmd` or dedicated helpers).
- Pass configuration explicitly (structs/params) instead of calling `os.Getenv` deep in call chains.
- Use `internal` helpers for session and config files (e.g., cookies, YAML config).

### Error handling and UX

- Commands should use `RunE` and return `error`.
- Wrap errors with context:
  - `fmt.Errorf("failed to <action>: %w", err)`.
- No panics in normal flow.
- Error messages should be direct and helpful (e.g., hint to run `interactiveai login` or `organizations select` when needed).
- When the server returns JSON with a `message` field, surface that string directly to the user.

### Testing and formatting

- Code must compile; add or update tests when behavior changes.
- Assume `go test ./...` should conceptually pass.
- All code must be `gofmt`-style; imports should follow Go conventions.

**Minimize the test code, not the test cases.** In a table-driven test the case
list can be as long as the behavior needs; the body of the loop should be a
call and one comparison.

- Assert one composite value rather than many fields. A whole request-target
  (`/…/skills/team%2Fdeploy?scope=global`) covers path, escaping and query
  params in one line, where separate checks on captured variables need one
  block each.
- Put `t.Cleanup` in the helper so no test carries its own `defer`.
- A test that passes whether or not the fix is present is not a test. Check by
  reverting the fix and watching it fail.

---

## 4. PR / change expectations

- Scope: one focused change (feature, bugfix, or small refactor).
- Description:
  - Mention commands and packages touched.
  - Summarize behavior changes.
  - Note any new flags, env vars, or config fields.
- Quality:
  - No unused code/imports.
  - Follows existing `cmd/` + `internal/` patterns.
  - Clear, minimal comments where truly necessary.

---

## 5. CLI verb and flag conventions

### Verbs

Use a consistent, minimal set of verbs across all resource commands:

| Verb | Alias | Purpose |
|------|-------|---------|
| `list` | `ls` | List/search/filter resources. **Never add a `search` verb** — filtering belongs on `list` via flags. |
| `get` | `describe`, `desc` | Retrieve a single resource by ID or name |
| `create` | — | Create a new resource |
| `update` | — | Update an existing resource |
| `delete` | `rm` | Delete a resource |
| `schema` | — | Display JSON schema (for agents, routines, policies, etc.) |

Do **not** introduce new verbs (e.g. `search`, `daily`, `inspect`, `show`) unless there is no way to express the functionality through the existing verb set with flags.

### Parent command aliases

Every parent command should have a singular alias:

```go
Use:     "observations",
Aliases: []string{"obs", "observation"},
```

### Flags

**Naming:**
- Hyphen-separated: `--trace-id`, `--user-id`, `--from-timestamp`
- Range filters: `--min-cost`, `--max-cost`, `--min-latency`, `--max-latency`
- Repeatable flags: plural form (`--tags`), use `StringArrayVar`
- Single-value filters: singular form (`--name`, `--model`, `--level`)

**Standard flags present on most commands:**
- `-o, --organization` — Organization name
- `-p, --project` — Project name
- `--json` — Output raw API response as JSON
- `--yaml` — Output raw API response as YAML
- `--columns` — Customize displayed table columns (`StringSliceVar`)

`--json` and `--yaml` are mutually exclusive with `--columns`; render the raw
payload with `output.PrintRawJSON` / `output.PrintRawYAML` and guard table-only
flags with `validateTableOnlyColumns`.

**Pagination (pick one per command, never both):**
- Page-based: `--page` (default 1), `--limit`
- Cursor-based: `--cursor`, `--limit`

**Timestamps:**
- List filters: `--from-timestamp` and `--to-timestamp` (ISO 8601 / RFC3339)
- Log windows: `--start-time` and `--end-time`, matching `services logs`

### Behavioral flags over new verbs

When a command has multiple modes or granularities, use a flag instead of a new subcommand:

```bash
# Good: flag controls granularity
iai metrics list --daily

# Bad: new verb for each granularity
iai metrics daily
iai metrics monthly
```

```bash
# Good: --trace-id scopes behavior on list
iai observations list --trace-id abc123
iai observations list --type GENERATION

# Bad: separate search command
iai observations search --type GENERATION
```

### Naming

Names are the interface. A consistent one is the difference between a CLI
someone can use without the docs and one they cannot.

- **One word per concept, everywhere.** The struct field, the JSON tag, the
  table column, the help text, the comments and the test fixtures all use it.
  Renaming a field and leaving the old word in the surrounding prose is the
  usual way this breaks.
- **Check what the API already calls it.** If the server takes `?scope=global`,
  the CLI reports `scope: global` — what a caller reads is what it passes back,
  with no mapping to learn.
- **Reuse a word only for the same question.** In this repo `type` answers
  *what kind of thing is this*, `scope` answers *whose namespace is this from*,
  `source` answers *what produced this record*. Reusing a familiar word for a
  different question is worse than introducing a new one.
- **Check what the word already means on the same object.** `PromptDetail`
  already carries `Type` and `RowType`; a third type-shaped field there would
  force readers to disambiguate three of them.
- **Leave the obvious name free** for the thing it obviously describes.

### Documentation

- Docs in `docs/` are markdown files named `iai_<resource>_<verb>.md`.
- When adding or renaming commands, update the parent doc's SEE ALSO section and create/rename the command doc accordingly.

---

## 6. Summary for agents

- Treat `interactive-cli` as a **thin, reliable client**:
  - Parse input → call backend → print response.
- Keep command code minimal; move real work to `internal/`.
- Follow the compact rules:
  - No commented-out code or unused imports.
  - Minimal, meaningful comments — shorter than the code they describe.
  - One word per concept, reused only for the same question.
  - Long test tables, short test bodies.
  - Light `init()`, fixed dependencies.
  - Clean error messages, no panics.
- Keep every change small, readable, and consistent with the existing style.