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
		name      string
		resource  string
		allowFlag string
		names     []string
		allowed   bool
		want      string
	}{
		{
			name:      "no deletions print nothing",
			resource:  "service",
			allowFlag: "services",
			names:     nil,
			allowed:   false,
			want:      "",
		},
		{
			name:      "no deletions print nothing even when allowed",
			resource:  "service",
			allowFlag: "services",
			names:     nil,
			allowed:   true,
			want:      "",
		},
		{
			name:      "refused single deletion names the unblocking flag",
			resource:  "agent",
			allowFlag: "agents",
			names:     []string{"chat-agent"},
			allowed:   false,
			want: "⚠ sync will NOT delete 1 agent not in the config: chat-agent" +
				" (a config that omits a resource looks identical to a stale one" +
				" — pass --allow-delete=agents to delete)\n",
		},
		{
			name:      "refused multiple deletions",
			resource:  "service",
			allowFlag: "services",
			names:     []string{"api-gateway", "worker"},
			allowed:   false,
			want: "⚠ sync will NOT delete 2 services not in the config: api-gateway, worker" +
				" (a config that omits a resource looks identical to a stale one" +
				" — pass --allow-delete=services to delete)\n",
		},
		{
			name:      "allowed single deletion",
			resource:  "agent",
			allowFlag: "agents",
			names:     []string{"chat-agent"},
			allowed:   true,
			want: "⚠ sync will DELETE 1 agent not in the config: chat-agent" +
				" (--allow-delete=agents; agent deletes run after agent creates/updates)\n",
		},
		{
			name:      "allowed multiple deletions",
			resource:  "service",
			allowFlag: "services",
			names:     []string{"api-gateway", "worker"},
			allowed:   true,
			want: "⚠ sync will DELETE 2 services not in the config: api-gateway, worker" +
				" (--allow-delete=services; service deletes run after service creates/updates)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintSyncDeletions(&buf, tt.resource, tt.allowFlag, tt.names, tt.allowed)
			if got := buf.String(); got != tt.want {
				t.Errorf("announcement = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintDroppedListEntries(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		flag        string
		live        []string
		incoming    []string
		want        string
		wantDropped bool
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
			wantDropped: true,
		},
		{
			name: "dropped secret ref",
			kind: "secret refs", flag: "--secret",
			live:     []string{"api-keys", "db-creds"},
			incoming: []string{"db-creds"},
			want: "⚠ this update drops live secret refs: api-keys" +
				" (--secret replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			dropped := PrintDroppedListEntries(&buf, tt.kind, tt.flag, tt.live, tt.incoming)
			if got := buf.String(); got != tt.want {
				t.Errorf("warning = %q, want %q", got, tt.want)
			}
			if dropped != tt.wantDropped {
				t.Errorf("dropped = %v, want %v", dropped, tt.wantDropped)
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
		name         string
		live         any
		incoming     any
		want         string
		wantBlocking bool
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
			wantBlocking: true,
		},
		{
			name: "removal is loud",
			live: configWithPins(nil, []any{
				map[string]any{"id": "handoff", "version": float64(4)},
			}, nil),
			incoming: configWithPins(nil, []any{}, nil),
			want: "⚠ content pins changed by this update:\n" +
				"    policy handoff: v4 → (none)  (REMOVED — this update drops the pin entirely)\n",
			wantBlocking: true,
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
			wantBlocking: true,
		},
		{
			name: "glossary and macro sections gate like the original three",
			live: map[string]any{"context": map[string]any{
				"glossaries": []any{
					map[string]any{"id": "billing-terms", "version": float64(6)},
				},
				"macros": []any{
					map[string]any{"id": "refund-flow", "version": float64(3)},
				},
			}},
			incoming: map[string]any{"context": map[string]any{
				"glossaries": []any{
					map[string]any{"id": "billing-terms", "version": 5},
				},
				"macros": []any{
					map[string]any{"id": "refund-flow", "version": 3},
				},
			}},
			want: "⚠ content pins changed by this update:\n" +
				"    glossary billing-terms: v6 → v5  (DOWNGRADE — if this is an intentional rollback, this is expected)\n",
			wantBlocking: true,
		},
		{
			name: "unknown pin-shaped section warns loudly but never blocks",
			live: map[string]any{"context": map[string]any{
				"playbooks": []any{
					map[string]any{"id": "escalation", "version": float64(9)},
				},
			}},
			incoming: map[string]any{"context": map[string]any{
				"playbooks": []any{
					map[string]any{"id": "escalation", "version": 8},
				},
			}},
			want: "⚠ content pins changed by this update:\n" +
				"    playbooks escalation: v9 → v8  (DOWNGRADE — if this is an intentional rollback, this is expected)\n",
			wantBlocking: false,
		},
		{
			name: "unknown section entries without a version key are not pins",
			live: map[string]any{"context": map[string]any{
				"tools": []any{
					map[string]any{"id": "github"},
					map[string]any{"id": "stripe"},
				},
			}},
			incoming: map[string]any{"context": map[string]any{
				"tools": []any{
					map[string]any{"id": "github"},
				},
			}},
			want: "",
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
			blocking := PrintPinChanges(&buf, tt.live, tt.incoming)
			if got := buf.String(); got != tt.want {
				t.Errorf("pin changes = %q, want %q", got, tt.want)
			}
			if blocking != tt.wantBlocking {
				t.Errorf("blocking = %v, want %v", blocking, tt.wantBlocking)
			}
		})
	}
}
