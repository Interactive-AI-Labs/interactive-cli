package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestRunUpdatePreflight(t *testing.T) {
	tests := []struct {
		name           string
		revision       int
		updated        string
		liveErr        error
		expectSet      bool
		expectRevision int
		wantOut        string
		wantErr        string
	}{
		{
			name:     "prints banner",
			revision: 12,
			updated:  "2026-07-23T16:40:00Z",
			wantOut:  "Live: revision 12, last updated 2026-07-23 16:40 UTC — this update creates revision 13\n",
		},
		{
			name:    "fails open on describe error",
			liveErr: errors.New("boom"),
			wantOut: "⚠ could not fetch live state (boom) — proceeding without pre-flight check\n",
		},
		{
			name:           "expect-revision match proceeds",
			revision:       12,
			expectSet:      true,
			expectRevision: 12,
			wantOut:        "Live: revision 12 — this update creates revision 13\n",
		},
		{
			name:           "expect-revision mismatch errors without banner",
			revision:       14,
			expectSet:      true,
			expectRevision: 12,
			wantOut:        "",
			wantErr:        "live revision is 14, expected 12 — not applying (--expect-revision)",
		},
		{
			name:           "expect-revision with describe error errors without fail-open note",
			liveErr:        errors.New("boom"),
			expectSet:      true,
			expectRevision: 12,
			wantOut:        "",
			wantErr:        "--expect-revision 12: could not verify live revision: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runUpdatePreflight(
				&buf, tt.revision, tt.updated, tt.liveErr, tt.expectSet, tt.expectRevision,
			)
			if got := buf.String(); got != tt.wantOut {
				t.Errorf("output = %q, want %q", got, tt.wantOut)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestPrintDroppedEnvSecretWarnings(t *testing.T) {
	liveEnv := []deployment.EnvVar{
		{Name: "LOG_LEVEL", Value: "info"},
		{Name: "DB_HOST", Value: "db"},
	}
	liveRefs := []deployment.SecretRef{{SecretName: "api-keys"}}

	tests := []struct {
		name          string
		envChanged    bool
		secretChanged bool
		envArgs       []string
		secretArgs    []string
		want          string
		wantDropped   bool
	}{
		{
			name:       "env replacement dropping a live var warns",
			envChanged: true,
			envArgs:    []string{"LOG_LEVEL=debug"},
			want: "⚠ this update drops live env vars: DB_HOST" +
				" (--env replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
		{
			name:          "secret replacement dropping a live ref warns",
			secretChanged: true,
			secretArgs:    []string{"other-secret"},
			want: "⚠ this update drops live secret refs: api-keys" +
				" (--secret replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
		{
			name:       "full list passed stays silent",
			envChanged: true,
			envArgs:    []string{"DB_HOST=db2", "LOG_LEVEL=debug", "EXTRA=1"},
			want:       "",
		},
		{
			name: "untouched flags stay silent",
			want: "",
		},
		{
			name:          "both lists dropping warn env first",
			envChanged:    true,
			secretChanged: true,
			envArgs:       []string{"NEW=1"},
			secretArgs:    []string{"other-secret"},
			want: "⚠ this update drops live env vars: DB_HOST, LOG_LEVEL" +
				" (--env replaces the entire list; pass every value you want to keep)\n" +
				"⚠ this update drops live secret refs: api-keys" +
				" (--secret replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			dropped := printDroppedEnvSecretWarnings(
				&buf,
				tt.envChanged, tt.secretChanged,
				liveEnv, liveRefs,
				tt.envArgs, tt.secretArgs,
			)
			if got := buf.String(); got != tt.want {
				t.Errorf("warnings = %q, want %q", got, tt.want)
			}
			if dropped != tt.wantDropped {
				t.Errorf("dropped = %v, want %v", dropped, tt.wantDropped)
			}
		})
	}
}

func TestAgentEnvWarnings(t *testing.T) {
	source := &deployment.EnvVarValueFrom{
		SecretKeyRef: deployment.SecretKeyRef{Name: "tools-dev", Key: "MCP_API_KEY"},
	}
	managed := deployment.EnvVar{Name: "MCP_KEY_TOOLS_DEV", ValueFrom: source}
	tests := []struct {
		name        string
		live        []deployment.EnvVar
		incoming    []string
		want        string
		wantDropped bool
	}{
		{
			name:     "adding user env does not warn about a preserved MCP reference",
			live:     []deployment.EnvVar{managed},
			incoming: []string{"LOG_LEVEL=debug"},
		},
		{
			name: "updating user env does not require resending the MCP reference",
			live: []deployment.EnvVar{
				{Name: "LOG_LEVEL", Value: "info"},
				managed,
			},
			incoming: []string{"LOG_LEVEL=debug"},
		},
		{
			name: "dropping user env still warns alongside managed MCP credentials",
			live: []deployment.EnvVar{
				managed,
				{Name: "DB_HOST", Value: "db"},
			},
			incoming: []string{"LOG_LEVEL=debug"},
			want: "⚠ this update drops live env vars: DB_HOST" +
				" (--env replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
		{
			name:     "MCP name without source metadata is not assumed managed",
			live:     []deployment.EnvVar{{Name: "MCP_KEY_TOOLS_DEV", Value: ""}},
			incoming: []string{"LOG_LEVEL=debug"},
			want: "⚠ this update drops live env vars: MCP_KEY_TOOLS_DEV" +
				" (--env replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
		{
			name:     "unmanaged secret references are not hidden",
			live:     []deployment.EnvVar{{Name: "CUSTOM_TOKEN", ValueFrom: source}},
			incoming: []string{"LOG_LEVEL=debug"},
			want: "⚠ this update drops live env vars: CUSTOM_TOKEN" +
				" (--env replaces the entire list; pass every value you want to keep)\n",
			wantDropped: true,
		},
		{
			name:     "no live env stays silent",
			incoming: []string{"LOG_LEVEL=debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			dropped := printDroppedEnvSecretWarnings(
				&buf, true, false, agentUserEnv(tt.live), nil, tt.incoming, nil,
			)
			if got := buf.String(); got != tt.want {
				t.Errorf("warnings = %q, want %q", got, tt.want)
			}
			if dropped != tt.wantDropped {
				t.Errorf("dropped = %v, want %v", dropped, tt.wantDropped)
			}
		})
	}
}

func TestCheckUpdateGates(t *testing.T) {
	tests := []struct {
		name           string
		force          bool
		pinRollback    bool
		droppedEntries bool
		wantErr        string
	}{
		{
			name: "nothing detected proceeds",
		},
		{
			name:        "pin rollback refuses",
			pinRollback: true,
			wantErr: "refusing to apply: this update downgrades or removes live content pins" +
				" (details above) — pass --force if this is intended",
		},
		{
			name:           "dropped entries refuse",
			droppedEntries: true,
			wantErr: "refusing to apply: this update drops live env vars or secret refs" +
				" (details above) — pass --force if this is intended",
		},
		{
			name:           "both reasons listed together",
			pinRollback:    true,
			droppedEntries: true,
			wantErr: "refusing to apply: this update downgrades or removes live content pins" +
				" and drops live env vars or secret refs" +
				" (details above) — pass --force if this is intended",
		},
		{
			name:           "force overrides everything",
			force:          true,
			pinRollback:    true,
			droppedEntries: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkUpdateGates(tt.force, tt.pinRollback, tt.droppedEntries)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
