# Agents guidelines for `interactive-cli`

This file defines how **AI coding agents** should work on the
`github.com/Interactive-AI-Labs/interactive-cli` Go/Cobra CLI.

The goal is to keep changes:
- Small and easy to review.
- Consistent with existing patterns.
- Safe to run in all environments.

---

## 1. Project context

- Language: **Go** (see `go.mod` for version; use that toolchain).
- Entry point: `main.go` → `cmd.Execute()`.
- CLI framework: **Cobra**.
- Responsibilities of this repo:
  - Authenticate to the Interactive AI platform.
  - Collect flags / arguments / environment.
  - Call HTTP APIs.
  - Render results (tables, messages, errors).
- **Do not** implement business logic or “agent brains” here. The platform is the source of truth; the CLI is a thin, reliable client.

When in doubt, prefer:
- “Forward input to the backend and print its response”
over
- “Rebuild backend behavior locally.”

---

## 2. How to work in this repo (as an AI coding agent)

When you receive a task:

1. **Understand the user request**
   - Identify which command(s) or feature(s) are affected.
   - Check existing commands in `cmd/` and helpers in `internal/` before adding new files.

2. **Locate relevant code**
   - Look for matching Cobra commands in `cmd/`.
   - Look for HTTP calls and shared helpers in `internal/`.
   - Prefer reusing existing patterns over inventing new ones.

3. **Plan a minimal change**
   - Avoid large refactors unless explicitly requested.
   - Keep each change focused (one feature or bugfix at a time).
   - Preserve existing public behavior unless the task explicitly says otherwise.

4. **Implement**
   - For new user-facing behavior:
     - Add or extend a Cobra command under `cmd/`.
     - Use small, testable functions in `internal/` to perform HTTP or formatting work.
   - Avoid unnecessary abstraction: one constructor and a few clear functions are usually enough.

5. **Validate**
   - Always ensure the code **compiles** and **tests pass** logically:
     - `go test ./...`
   - Where applicable, consider:
     - `go vet ./...` for basic static checks.
     - Running only affected packages (e.g. `go test ./cmd/...`).

6. **Document minimally**
   - Update comments or README-like docs only when they add real value.
   - Prefer clear names and self-explanatory code instead of verbose commentary.

---

## 3. Code structure rules

### 3.1 Cobra commands vs internal logic

- `cmd/` package:
  - Contains Cobra commands and CLI wiring only.
  - Responsibilities:
    - Command usage strings, descriptions, and flags.
    - Input parsing (arguments and flags).
    - High-level user messaging.
    - Delegation into `internal/` for actual work.
- `internal/` package:
  - Contains reusable helpers for HTTP calls, formatting, storage, and other utilities.
  - Responsibilities:
    - Constructing URLs and HTTP requests.
    - Handling responses and errors.
    - Data formatting (tables, text).

**Guideline for AI agents:**
- Do not put complex logic inside `Run` / `RunE` bodies.
- Put that logic into small functions or structs in `internal/` and call them from the command.

### 3.2 Locality of behavior

- Keep related behavior together:
  - HTTP client helpers and their data types should live in the same file or package.
  - A Cobra command should either:
    - Call `internal` functions directly, or
    - Use a small helper in the same `internal` package that hides HTTP details.
- Avoid spreading one logical workflow across many packages unless there is a strong reason.

---

## 4. Go coding conventions

These adapt general Python guidelines to Go for this repo.

### 4.1 Never commit commented-out code

- Do **not** leave old implementations as comments.
- If code is unused, **delete it**.
- If you may need it later, rely on Git history, not comments.

### 4.2 Only commit comments that are strictly necessary

- Prefer expressive names for:
  - Types, functions, methods.
  - Variables and fields.
- Comments should:
  - Explain **why**, not **what**.
  - Document non-obvious contracts, edge cases, or constraints.
  - Describe interactions with external APIs when behavior is surprising.
- Avoid doc comments that merely restate function signatures.

### 4.3 Don’t keep unused code or imports

- Remove functions, types, and helpers when they are no longer used.
- Keep import lists minimal and free of unused entries.
- Avoid hacks like `_ = someName` just to silence “unused” warnings.

### 4.4 Dependencies must be fixed

- Use the `go.mod` file; depend on explicit versions.
- Avoid “floating” or ambiguous versions.
- When adding or updating a dependency:
  - Keep changes minimal.
  - Prefer libraries already in use in the org, if applicable.
  - Only add a new dependency when the standard library and existing deps are clearly insufficient.

### 4.5 Avoid heavy work at import time

- Keep `init()` functions light and predictable.
- Do **not**:
  - Start goroutines.
  - Open files.
  - Create network clients that make requests.
- Prefer constructors and explicit wiring in Cobra commands.

### 4.6 Environment variables

- Read core environment variables in a **small number of central places** (usually near the root command or main).
- Pass configuration into helpers explicitly (e.g. via structs or parameters).
- Avoid sprinkling `os.Getenv` throughout deep call chains for the same variable.
- Prefer explicit configuration for testability and clarity.

### 4.7 Avoid unnecessary boilerplate

- Do not introduce extra wrapper functions that:
  - Only pass parameters through without adding behavior or clarity.
  - Force extra layers of indirection to understand what the code does.
- If a constructor or helper is simple and used in one place, keep it simple and local.
- Introduce abstractions **only** when:
  - They remove real duplication, or
  - They clearly improve readability / testability.

---

## 5. Error handling and UX

- All public commands:
  - Should return `error` from `RunE` so Cobra can handle failure exit codes cleanly.
  - Should provide clear, actionable error messages for the user.
- Wrap errors with context:
  - Use `fmt.Errorf("failed to <action>: %w", err)` patterns.
  - Make it obvious what the CLI was trying to do when the failure happened.
- Avoid panics in normal control flow.
- Prefer:
  - Simple, consistent messaging.
  - Helpful hints (e.g. “Please run `<cli> login` first” when session is missing).

---

## 6. Testing and formatting

- Before considering a change “done”, AI coding agents should:
  - Ensure the code **compiles**.
  - Add or update **tests** in affected packages when behavior changes.
- Typical commands to consider (conceptually):
  - `go test ./...` – run all tests.
  - `go test ./cmd/...` – focus on CLI wiring tests (if present).
  - `go test ./internal/...` – focus on helpers and HTTP logic.
- Formatting:
  - All Go code should be formatted with `gofmt` style (most editors / tools do this automatically).
  - Keep import groupings consistent with standard Go conventions.

---

## 7. PR and change guidelines (for AI agents)

When generating changes that will become a PR:

1. **Scope**
   - Keep the change focused on a single feature, bugfix, or refactor.
   - Avoid mixing unrelated edits.

2. **Title and description**
   - Use a concise, descriptive title (e.g. `[interactive-cli] Add projects list command`).
   - In the description:
     - Name the commands and packages touched.
     - Summarize behavior changes.
     - Mention any new environment variables or configuration expectations.

3. **Quality checks**
   - Code is compiled and logically passes tests.
   - No unused code or imports.
   - Comments, if present, are minimal and necessary.
   - New code follows existing patterns in `cmd/` and `internal/`.

---

## 8. Summary for AI coding agents

When you work on `interactive-cli`:

- Treat this repo as a **thin client** to the Interactive AI platform.
- Keep commands small and delegate real work to `internal/` helpers.
- Respect the coding rules:
  - No commented-out code.
  - Minimal, meaningful comments.
  - No unused code or imports.
  - Fixed dependencies and light `init()` behavior.
- Add tests for new behavior.
- Keep changes scoped, readable, and aligned with existing code.

Following these rules will keep the CLI maintainable and predictable for humans and AI agents alike.
