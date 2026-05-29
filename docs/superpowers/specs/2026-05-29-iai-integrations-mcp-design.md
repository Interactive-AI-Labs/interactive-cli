# `iai integrations` — MCP integration connections in the CLI

**Date:** 2026-05-29
**Repo:** `interactive-cli`
**Status:** Approved design, ready for implementation plan

## Problem

The platform recently shipped a feature to connect integrations (MCP servers) to
a project: browse a catalog, create custom or catalog-backed connections, verify
them, inspect their tools, and run a tool. The full HTTP surface exists under
`/api/platform/v1/.../mcp-connections` and `/.../mcp-catalog`.

The `iai` CLI cannot drive any of it. As a consequence the interactive copilot —
which operates the platform exclusively through `iai` commands run in its
sandbox — also cannot manage integrations.

This work extends the CLI with an `iai integrations` command group covering the
full connection lifecycle plus tool execution, mirroring the existing CLI
conventions exactly. Publishing the regenerated CLI docs is in scope. **Wiring
the copilot to discover the new commands (sandbox image pin bump, `fetch_docs`
allowlist, `DOC_TOPICS` map) is explicitly out of scope** and will be a separate
follow-up in `interactive-mcp` / `interactive-chat`.

## Scope

**In scope (this work, `interactive-cli` only):**
- New `iai integrations` command group with eight commands (below).
- New `APIClient` methods for the seven backend endpoints.
- New output printers for connection list/detail, catalog, verify, and tool-run.
- Regenerated `docs/` via `make docs`, committed.
- Unit tests following the existing `*_test.go` patterns.

**Out of scope (separate follow-up):**
- `interactive-mcp/Dockerfile.sandbox` version pin bump.
- `interactive-chat/agent/tools.py` `fetch_docs` URL list.
- `interactive-mcp/agent/docs.py` `DOC_TOPICS` map.
- Publishing `docs.interactive.ai/cli/iai_integrations*.md` to the docs site
  (the CLI repo only regenerates its local `docs/`; site publication is a
  separate pipeline).

## Backend contract (verified against `interactive-platform`)

All endpoints are under prefix `/api/platform/v1` and include both org and
project IDs in the path. Responses are wrapped in the standard
`{ "success": true, "data": { ... } }` envelope.

Base: `/api/platform/v1/organizations/{org}/projects/{project}`

| Op | Method + path (relative to base) | Request body | `data` shape |
|---|---|---|---|
| List connections | `GET /mcp-connections` | — | `{ connections: McpConnection[] }` |
| Create | `POST /mcp-connections` | `McpConnectionCreateRequest` | `{ connection: McpConnectionDetail }` |
| Get detail | `GET /mcp-connections/{id}` | — | `{ connection: McpConnectionDetail }` |
| Delete | `DELETE /mcp-connections/{id}` | — | `{ id, deleted }` |
| Verify | `POST /mcp-connections/{id}/verify` | — | `McpVerifyData` |
| Run tool | `POST /mcp-connections/{id}/tools/{tool_name}/run` | `{ arguments: object }` | `McpToolCallData` |
| Catalog | `GET /mcp-catalog` | — | `{ entries: McpCatalogEntry[] }` |

### Create request (`McpConnectionCreateRequest`)

```
type:           "platform" | "custom"          (required)
catalog_id:     string                          (required iff type=platform)
name:           string  (1..128)                (required)
slug:           string  (<=64, optional; tool prefix, auto-derived from name)
description:    string  (optional)
endpoint_url:   string                          (required)
transport:      "streamable_http" | "sse"       (default streamable_http)
auth_type:      "api_key" | "bearer" | "none"   (required)
credential:     string  (required unless auth_type=none; stored encrypted)
custom_headers: { string: string }              (optional)
```

The backend enforces: `catalog_id` required for `platform`, forbidden for
`custom`. Create verifies the server on save and rolls back (502) on dial
failure.

### Read shapes (credential never returned)

`McpConnection`: `id, project_id, catalog_id?, type, name, slug, description?,
endpoint_url, transport, auth_type, has_credential, custom_headers, status
(pending|ok|error), last_verified_at?, last_error_class?, protocol_version?,
tool_count, connected_agents[{id,name}], created_by, created_at, updated_at`.

`McpConnectionDetail` = `McpConnection` + `tools: McpTool[]` + `last_error?`.

`McpTool`: `name, description?, input_schema?, enabled`.

`McpVerifyData`: `status, error_class?, error_message?, protocol_version?,
server_info?, tools[]`.

`McpToolCallData`: `status (ok|error), result?, error_class?, error_message?`.

`McpCatalogEntry`: `id, name, category, description, type, icon_key,
endpoint_url?, docs_url?, auth_methods[]`.

### Auth

The CLI's existing `APIClient` (Basic-Auth API key or session cookies, resolved
by `ApplyAuth`) is the correct client. SDK/Basic-Auth mode requires the org and
project in the path to match the key's project — `resolveProject` already yields
both `orgId` and `projectId`, so no new auth work is needed.

## Command surface

New group registered under `GroupID: groupInfra` (Infrastructure), matching
where connections conceptually sit alongside services/secrets/databases.

```
iai integrations list
iai integrations get <id>
iai integrations create-custom <name>
iai integrations create-from-catalog <name>
iai integrations delete <id>
iai integrations verify <id>
iai integrations catalog
iai integrations tools run <id> <tool>
```

Parent command:

```go
parentCmd := &cobra.Command{
    Use:     "integrations",
    Aliases: []string{"integration", "mcp"},
    Short:   "MCP integration connections for a project",
    GroupID: groupInfra,
    Long:    `Manage Model Context Protocol (MCP) integration connections ...`,
}
```

Every leaf command takes the standard `-p/--project` and `-o/--organization`
flags and calls `resolveProject(cmd.Context(), org, project)`, exactly like
`prompts`/`secrets`.

### Why two `create` commands instead of one

Decision: split create into `create-custom` and `create-from-catalog` rather
than a single `create` that infers type from flags.

Rationale — optimized for copilot usability (the copilot constructs raw `iai`
strings from per-command doc pages):
- Named commands map 1:1 to user intent; no conditional flag-logic for the model
  to misapply.
- Each command's required flags are unconditional, eliminating
  invalid-combination round-trips (e.g. `--from-catalog` + `--endpoint-url`).
- `gen-docs` emits one focused page per command, which is easier to consume than
  one page hedging two mutually-exclusive flag sets.

This is the faithful translation of the backend's explicit `type` discriminator,
not a departure from sibling commands (no sibling has this dual-shape problem).

### `create-custom <name>`

Sends `type="custom"` (no `catalog_id`).

| Flag | Required | Maps to |
|---|---|---|
| `--endpoint-url` | yes | `endpoint_url` |
| `--auth-type` | yes | `auth_type` (`api_key`\|`bearer`\|`none`) |
| `--credential` | iff auth-type != none | `credential` |
| `--transport` | no (default `streamable_http`) | `transport` |
| `--slug` | no | `slug` |
| `--description` | no | `description` |
| `--header KEY=VALUE` (repeatable, `StringArray`) | no | `custom_headers` |

Client-side validation mirrors the server: reject `none` + `--credential`, and
require `--credential` when auth-type is `api_key`/`bearer`, with a clear error
before the request is sent. Validate `--auth-type`/`--transport` against their
enums (same pattern as `validatePromptType`).

### `create-from-catalog <name>`

Sends `type="platform"` with `catalog_id`.

| Flag | Required | Maps to |
|---|---|---|
| `--catalog-id` | yes | `catalog_id` |
| `--auth-type` | yes | `auth_type` |
| `--credential` | iff auth-type != none | `credential` |
| `--slug` | no | `slug` |
| `--description` | no | `description` |

`endpoint_url`/`transport` come from the catalog entry server-side; the CLI does
not send them for catalog-backed connections.

### `delete <id>`

Destructive → confirmation prompt unless `-f/--force`, exactly like
`prompts delete`.

### `tools run <id> <tool>`

`tools` is an intermediate parent command (like `agents revisions`), `run` is the
leaf. Arguments to the tool:

| Flag | Maps to |
|---|---|
| `--args '<json>'` | parsed as a JSON object → `arguments` |
| `--args-file <path>` | file contents parsed as a JSON object → `arguments` |

`--args` and `--args-file` are mutually exclusive (`MarkFlagsMutuallyExclusive`).
Default is `{}`. The CLI parses+validates the JSON is an object before sending,
emitting a clear error otherwise.

### `verify <id>` / `catalog` / `list` / `get`

No special flags beyond project/org. `catalog` and `list` print tables; `get`
prints a detail view including the tool list; `verify` prints the verify result
(status + protocol/server info on success, error class/message on failure).

## Client methods (`internal/clients/api_client.go`)

Use the existing generics (`doGet[T]`, `doCreate[T]`, `doList[T]`, `doDelete`)
and `decodeSuccess[T]` — the modern idiom already used by traces/observations/
scores. New typed structs (Go-side mirrors of the read shapes) live alongside.

```go
func (c *APIClient) ListMcpConnections(ctx, orgId, projectId) (*McpConnectionListData, error)
func (c *APIClient) GetMcpConnection(ctx, orgId, projectId, id) (*McpConnectionDetail, error)
func (c *APIClient) CreateMcpConnection(ctx, orgId, projectId, body McpConnectionCreateBody) (*McpConnectionDetail, error)
func (c *APIClient) DeleteMcpConnection(ctx, orgId, projectId, id) error
func (c *APIClient) VerifyMcpConnection(ctx, orgId, projectId, id) (*McpVerifyData, error)
func (c *APIClient) RunMcpTool(ctx, orgId, projectId, id, tool, args map[string]any) (*McpToolCallData, error)
func (c *APIClient) ListMcpCatalog(ctx, orgId, projectId) (*McpCatalogListData, error)
```

A small `mcpBasePath(orgId, projectId)` helper builds the
`/api/platform/v1/organizations/{org}/projects/{project}` prefix with
`url.PathEscape`, mirroring `promptBasePath`. `tool_name` in the run path is
path-escaped.

`CreateMcpConnection` takes a single body struct; the two `create-*` commands
populate it differently (type + which fields they set) and share this one method.

## Output (`internal/output/integrations.go`)

New printers using the existing `PrintTable` / `NewDescribeWriter` /
`TruncateList` / `LocalTime` helpers — no new formatting primitives.

- `PrintMcpConnectionList`: table `NAME | TYPE | STATUS | TOOLS | ENDPOINT | UPDATED`.
- `PrintMcpConnectionDetail`: describe-writer block (name, id, type, status,
  slug, endpoint, transport, auth type, protocol version, connected agents,
  last verified/error) followed by a `TOOLS` sub-table (`NAME | ENABLED |
  DESCRIPTION`).
- `PrintMcpCatalog`: table `ID | NAME | CATEGORY | TYPE | AUTH`.
- `PrintMcpVerifyResult`: status line + protocol/server info on ok, error
  class/message on error, then the refreshed tool table.
- `PrintMcpToolResult`: status + pretty-printed JSON `result`, or error
  class/message.

Status values (`ok`/`error`/`pending`) get the same coloring treatment used
elsewhere if a sibling already colorizes status; otherwise plain text.

## Error handling

- HTTP/envelope errors surface through the existing `ExtractServerMessage` /
  `decodeSuccess` path — the user sees the backend's message (e.g. the 502
  "connection failed on save" or 404 "tool not found / disabled").
- Client-side pre-flight validation (auth-type/transport enums, credential
  presence, JSON-object args) fails fast with a clear message before any request.
- `create-*` surface the verify-on-save failure verbatim; the connection is not
  created server-side in that case (backend rolls back), so no special cleanup.

## Docs

After implementation, run `make docs` (→ `go run main.go gen-docs`) to
regenerate the per-command Markdown under `docs/` (new files:
`iai_integrations.md`, `iai_integrations_list.md`, `..._get.md`,
`..._create-custom.md`, `..._create-from-catalog.md`, `..._delete.md`,
`..._verify.md`, `..._catalog.md`, `..._tools.md`, `..._tools_run.md`). Commit
the regenerated docs in the same PR.

## Testing

Follow existing `*_test.go` conventions:
- `internal/output/integrations_test.go`: golden-style assertions on each printer
  (empty list, populated list, detail with/without tools, verify ok/error, tool
  result ok/error), matching `prompts_test.go`/`secrets_test.go`.
- `internal/clients/api_client_test.go` additions: `httptest` server asserting
  method + path + body for each new method, and envelope decoding (incl.
  `success=false` → error).
- Command-level: validation-error tests (missing required flags, bad enum,
  `none`+credential, non-object `--args`) where a sibling command has analogous
  tests.

## File plan

| File | Change |
|---|---|
| `cmd/integrations.go` | new — command group + 8 commands |
| `internal/clients/api_client.go` | new methods + `mcpBasePath` + typed structs |
| `internal/output/integrations.go` | new — printers |
| `internal/output/integrations_test.go` | new — printer tests |
| `internal/clients/api_client_test.go` | add MCP method tests |
| `docs/iai_integrations*.md` | regenerated via `make docs` |

No changes to `root.go` (reusing `groupInfra`).
