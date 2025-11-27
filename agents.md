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

## 5. Summary for agents

- Treat `interactive-cli` as a **thin, reliable client**:
  - Parse input → call backend → print response.
- Keep command code minimal; move real work to `internal/`.
- Follow the compact rules:
  - No commented-out code or unused imports.
  - Minimal, meaningful comments.
  - Light `init()`, fixed dependencies.
  - Clean error messages, no panics.
- Keep every change small, readable, and consistent with the existing style.