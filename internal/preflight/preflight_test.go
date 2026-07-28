package preflight

import (
	"bytes"
	"errors"
	"testing"
)

func TestPrintUpdateBanner(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		revision int
		updated  string
		want     string
	}{
		{
			name:     "with timestamp",
			revision: 12,
			updated:  "2026-07-23T16:40:00Z",
			want:     "Live: revision 12, last updated 2026-07-23 16:40 UTC — this update creates revision 13\n",
		},
		{
			name:     "timestamp normalized to UTC",
			revision: 12,
			updated:  "2026-07-23T17:40:00+01:00",
			want:     "Live: revision 12, last updated 2026-07-23 16:40 UTC — this update creates revision 13\n",
		},
		{
			name:     "unparseable timestamp printed raw",
			revision: 3,
			updated:  "last Tuesday",
			want:     "Live: revision 3, last updated last Tuesday — this update creates revision 4\n",
		},
		{
			name:     "missing timestamp elided",
			revision: 3,
			want:     "Live: revision 3 — this update creates revision 4\n",
		},
		{
			name:     "named target",
			target:   "agent notify-agent-dev",
			revision: 13,
			updated:  "2026-07-24T11:20:00Z",
			want:     "Live: agent notify-agent-dev revision 13, last updated 2026-07-24 11:20 UTC — this update creates revision 14\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintUpdateBanner(&buf, tt.target, tt.revision, tt.updated)
			if got := buf.String(); got != tt.want {
				t.Errorf("banner = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintFailOpenNote(t *testing.T) {
	var buf bytes.Buffer
	PrintFailOpenNote(&buf, "fetch live state", errors.New("connection refused"))
	want := "⚠ could not fetch live state (connection refused) — proceeding without pre-flight check\n"
	if got := buf.String(); got != want {
		t.Errorf("note = %q, want %q", got, want)
	}
}

func TestPrintTagOverwriteWarning(t *testing.T) {
	var buf bytes.Buffer
	PrintTagOverwriteWarning(&buf, "0.2.36")
	want := "⚠ tag 0.2.36 already exists upstream — pushing replaces it; the previous image is unrecoverable.\n"
	if got := buf.String(); got != want {
		t.Errorf("warning = %q, want %q", got, want)
	}
}

func TestPrintSyncDeletions(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		names    []string
		want     string
	}{
		{
			name:     "no deletions print nothing",
			resource: "service",
			names:    nil,
			want:     "",
		},
		{
			name:     "single deletion",
			resource: "agent",
			names:    []string{"chat-agent"},
			want: "⚠ sync will DELETE 1 agent not in the config: chat-agent" +
				" (deletes run last — Ctrl-C to abort if the config is stale)\n",
		},
		{
			name:     "multiple deletions",
			resource: "service",
			names:    []string{"api-gateway", "worker"},
			want: "⚠ sync will DELETE 2 services not in the config: api-gateway, worker" +
				" (deletes run last — Ctrl-C to abort if the config is stale)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintSyncDeletions(&buf, tt.resource, tt.names)
			if got := buf.String(); got != tt.want {
				t.Errorf("announcement = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintDroppedListEntries(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		flag     string
		live     []string
		incoming []string
		want     string
	}{
		{
			name: "nothing dropped prints nothing",
			kind: "env vars", flag: "--env",
			live:     []string{"LOG_LEVEL"},
			incoming: []string{"LOG_LEVEL", "NEW_VAR"},
			want:     "",
		},
		{
			name: "empty live list prints nothing",
			kind: "env vars", flag: "--env",
			live:     nil,
			incoming: []string{"LOG_LEVEL"},
			want:     "",
		},
		{
			name: "dropped env vars listed sorted",
			kind: "env vars", flag: "--env",
			live:     []string{"LOG_LEVEL", "DB_HOST"},
			incoming: []string{"NEW_VAR"},
			want: "⚠ this update drops live env vars: DB_HOST, LOG_LEVEL" +
				" (--env replaces the entire list; pass every value you want to keep)\n",
		},
		{
			name: "dropped secret ref",
			kind: "secret refs", flag: "--secret",
			live:     []string{"api-keys", "db-creds"},
			incoming: []string{"db-creds"},
			want: "⚠ this update drops live secret refs: api-keys" +
				" (--secret replaces the entire list; pass every value you want to keep)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintDroppedListEntries(&buf, tt.kind, tt.flag, tt.live, tt.incoming)
			if got := buf.String(); got != tt.want {
				t.Errorf("warning = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckExpectedRevision(t *testing.T) {
	if err := CheckExpectedRevision(12, 12); err != nil {
		t.Errorf("matching revisions: unexpected error %v", err)
	}
	err := CheckExpectedRevision(12, 14)
	if err == nil {
		t.Fatal("expected error on revision mismatch")
	}
	want := "live revision is 14, expected 12 — not applying (--expect-revision)"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// configWithPins builds an agent config with the context shapes the CLI
// sees: routines/policies as lists of {id, version}, system_prompt as a
// single {id, version} object.
func configWithPins(systemPrompt map[string]any, policies, routines []any) map[string]any {
	ctx := map[string]any{}
	if systemPrompt != nil {
		ctx["system_prompt"] = systemPrompt
	}
	if policies != nil {
		ctx["policies"] = policies
	}
	if routines != nil {
		ctx["routines"] = routines
	}
	return map[string]any{"context": ctx}
}

func TestPrintPinChanges(t *testing.T) {
	tests := []struct {
		name     string
		live     any
		incoming any
		want     string
	}{
		{
			name: "downgrade is loud with rollback-aware wording",
			live: configWithPins(nil, nil, []any{
				map[string]any{"id": "notify-welcome-dev", "version": float64(13)},
			}),
			incoming: configWithPins(nil, nil, []any{
				map[string]any{"id": "notify-welcome-dev", "version": 12},
			}),
			want: "⚠ content pins changed by this update:\n" +
				"    routine notify-welcome-dev: v13 → v12  (DOWNGRADE — if this is an intentional rollback, this is expected)\n",
		},
		{
			name: "removal is loud",
			live: configWithPins(nil, []any{
				map[string]any{"id": "handoff", "version": float64(4)},
			}, nil),
			incoming: configWithPins(nil, []any{}, nil),
			want: "⚠ content pins changed by this update:\n" +
				"    policy handoff: v4 → (none)  (REMOVED — this update drops the pin entirely)\n",
		},
		{
			name: "upgrade and addition are quiet",
			live: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": float64(12)},
			}),
			incoming: configWithPins(
				map[string]any{"id": "base-prompt", "version": 2},
				nil,
				[]any{map[string]any{"id": "welcome", "version": 13}},
			),
			want: "content pins changed by this update:\n" +
				"    routine welcome: v12 → v13\n" +
				"    system_prompt base-prompt: (none) → v2\n",
		},
		{
			name: "unchanged pins print nothing",
			live: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": float64(13)},
			}),
			incoming: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": 13},
			}),
			want: "",
		},
		{
			name: "json float and yaml int compare equal",
			live: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": float64(13)},
			}),
			incoming: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": "13"},
			}),
			want: "",
		},
		{
			name: "non-numeric version change is quiet",
			live: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": "2026-06-01.a"},
			}),
			incoming: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": "2026-07-01.b"},
			}),
			want: "content pins changed by this update:\n" +
				"    routine welcome: v2026-06-01.a → v2026-07-01.b\n",
		},
		{
			name: "mixed delta sorts by section then id and flags only the loud lines",
			live: configWithPins(nil,
				[]any{map[string]any{"id": "guard", "version": float64(3)}},
				[]any{
					map[string]any{"id": "goodbye", "version": float64(8)},
					map[string]any{"id": "welcome", "version": float64(12)},
				},
			),
			incoming: configWithPins(nil,
				[]any{map[string]any{"id": "guard", "version": 3}},
				[]any{
					map[string]any{"id": "goodbye", "version": 7},
					map[string]any{"id": "welcome", "version": 13},
				},
			),
			want: "⚠ content pins changed by this update:\n" +
				"    routine goodbye: v8 → v7  (DOWNGRADE — if this is an intentional rollback, this is expected)\n" +
				"    routine welcome: v12 → v13\n",
		},
		{
			name: "unpinned entries render as (unpinned)",
			live: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome", "version": float64(2)},
			}),
			incoming: configWithPins(nil, nil, []any{
				map[string]any{"id": "welcome"},
			}),
			want: "content pins changed by this update:\n" +
				"    routine welcome: v2 → (unpinned)\n",
		},
		{
			name:     "non-map configs print nothing",
			live:     "not a map",
			incoming: 42,
			want:     "",
		},
		{
			name: "config without context prints nothing",
			live: map[string]any{"mcps": []any{"github"}},
			incoming: map[string]any{
				"mcps": []any{"github", "stripe"},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintPinChanges(&buf, tt.live, tt.incoming)
			if got := buf.String(); got != tt.want {
				t.Errorf("pin changes = %q, want %q", got, tt.want)
			}
		})
	}
}
